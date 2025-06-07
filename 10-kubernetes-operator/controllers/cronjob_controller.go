/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/robfig/cron"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "slurm.io/cronjob/api/v1"
)

// CronJobReconciler анализирует обьект CronJob
type CronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

// Clock знают как получить текущее время
// Пригодится нас для тестирования, чтобы заинжектить какие нить свои часики.
type Clock interface {
	Now() time.Time
}

//+kubebuilder:rbac:groups=batch.slurm.io,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.slurm.io,resources=cronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.slurm.io,resources=cronjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

var (
	scheduledTimeAnnotation = "batch.cronjob.slurm.io/scheduled-at"
)

func isJobFinished(job *kbatch.Job) (bool, kbatch.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}

	return false, ""
}

func getScheduledTimeForJob(job *kbatch.Job) (*time.Time, error) {
	timeRaw := job.Annotations[scheduledTimeAnnotation]
	if len(timeRaw) == 0 {
		return nil, nil
	}

	timeParsed, err := time.Parse(time.RFC3339, timeRaw)
	if err != nil {
		return nil, err
	}
	return &timeParsed, nil
}

func getNextSchedule(cronJob *batchv1.CronJob, now time.Time) (lastMissed time.Time, next time.Time, err error) {
	sched, err := cron.ParseStandard(cronJob.Spec.Schedule)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Unparseable schedule %q: %v", cronJob.Spec.Schedule, err)
	}

	// для оптимизации немного считерим и стартанем с последнего времени запуска
	// можно было бы восстановить его, но нет смысла, поскольку мы ж его только что и апдейтили
	var earliestTime time.Time
	if cronJob.Status.LastScheduleTime != nil {
		earliestTime = cronJob.Status.LastScheduleTime.Time
	} else {
		earliestTime = cronJob.ObjectMeta.CreationTimestamp.Time
	}
	if cronJob.Spec.StartingDeadlineSeconds != nil {
		// контроллер не создаст джобу за этой временной точкой
		schedulingDeadline := now.Add(-time.Second * time.Duration(*cronJob.Spec.StartingDeadlineSeconds))

		if schedulingDeadline.After(earliestTime) {
			earliestTime = schedulingDeadline
		}
	}
	if earliestTime.After(now) {
		return time.Time{}, sched.Next(now), nil
	}

	starts := 0
	for t := sched.Next(earliestTime); !t.After(now); t = sched.Next(t) {
		lastMissed = t
		// Мы можем пропустить несколько запусков. Например, если
		// контроллер подвис в пятницу в 17:01, когда все уже бухают
		// и кто то пришел во вторник и перезапустил контроллер
		// и все часовые работы, которых щас должно быть в районе 80 для одной  scheduledJob,
		// должны все запуститься без каких либо вопросов (если конечно мы разрешили конкурентность)
		// и поздний старт
		//
		// Тем не менее если где то баг, или некорректные часы
		// на сервере контроллера или где то в апи
		// мы можем запустить тысячи джобов и повесить нашу систему, поэтому
		// тут должно быть разумное ограничение.
		starts++
		if starts > 100 {
			// Мы не можем получить даты, поэтому просто вернем пустое Time
			return time.Time{}, time.Time{}, fmt.Errorf("Too many missed start times (> 100). Set or decrease .spec.startingDeadlineSeconds or check clock skew.")
		}
	}
	return lastMissed, sched.Next(now), nil
}

func constructJobForCronJob(cronJob *batchv1.CronJob, scheduledTime time.Time, scheme *runtime.Scheme) (*kbatch.Job, error) {
	// Нам нужно имя джоба с номинальным стартовым временем, чтобы их отличать друг от друга и не создавать одно и то же дважды
	name := fmt.Sprintf("%s-%d", cronJob.Name, scheduledTime.Unix())

	job := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   cronJob.Namespace,
		},
		Spec: *cronJob.Spec.JobTemplate.Spec.DeepCopy(),
	}
	for k, v := range cronJob.Spec.JobTemplate.Annotations {
		job.Annotations[k] = v
	}
	job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)
	for k, v := range cronJob.Spec.JobTemplate.Labels {
		job.Labels[k] = v
	}
	if err := ctrl.SetControllerReference(cronJob, job, scheme); err != nil {
		return nil, err
	}

	return job, nil
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var cronJob batchv1.CronJob
	if err := r.Get(ctx, req.NamespacedName, &cronJob); err != nil {

		log.Log.Error(err, "unable to fetch CronJob")
		// мы проигнорируем ошибку "не найдено" поскольку их нельзя разрешить немедленным
		// перезапуском (нам нужно дождаться новой нотификации)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var childJobs kbatch.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		log.Log.Error(err, "unable to list child Jobs")
		return ctrl.Result{}, err
	}
	// найти активные джобы
	var activeJobs []*kbatch.Job
	var successfulJobs []*kbatch.Job
	var failedJobs []*kbatch.Job
	var mostRecentTime *time.Time // понять когда был последний запуск чтобы можно было проапдейтить статус

	for i, job := range childJobs.Items {
		_, finishedType := isJobFinished(&job)
		switch finishedType {
		case "": // активные джобы
			activeJobs = append(activeJobs, &childJobs.Items[i])
		case kbatch.JobFailed:
			failedJobs = append(failedJobs, &childJobs.Items[i])
		case kbatch.JobComplete:
			successfulJobs = append(successfulJobs, &childJobs.Items[i])
		}

		// Мы будем держать время в аннотации, чтобы восстановить для
		// для активных джобов.
		scheduledTimeForJob, err := getScheduledTimeForJob(&job)
		if err != nil {
			log.Log.Error(err, "unable to parse schedule time for child job", "job", &job)
			continue
		}
		if scheduledTimeForJob != nil {
			if mostRecentTime == nil {
				mostRecentTime = scheduledTimeForJob
			} else if mostRecentTime.Before(*scheduledTimeForJob) {
				mostRecentTime = scheduledTimeForJob
			}
		}
	}

	if mostRecentTime != nil {
		cronJob.Status.LastScheduleTime = &metav1.Time{Time: *mostRecentTime}
	} else {
		cronJob.Status.LastScheduleTime = nil
	}
	cronJob.Status.Active = nil
	for _, activeJob := range activeJobs {
		jobRef, err := ref.GetReference(r.Scheme, activeJob)
		if err != nil {
			log.Log.Error(err, "unable to make reference to active job", "job", activeJob)
			continue
		}
		cronJob.Status.Active = append(cronJob.Status.Active, *jobRef)
	}
	log.Log.V(1).Info("job count", "active jobs", len(activeJobs), "successful jobs", len(successfulJobs), "failed jobs", len(failedJobs))
	if err := r.Status().Update(ctx, &cronJob); err != nil {
		log.Log.Error(err, "unable to update CronJob status")
		return ctrl.Result{}, err
	}

	// NB: Удаление это "best effort" -- если мы провалимся, нам не надо перезапускать,
	// чтобы просто завершить удаление, закончим в другой раз.
	if cronJob.Spec.FailedJobsHistoryLimit != nil {
		sort.Slice(failedJobs, func(i, j int) bool {
			if failedJobs[i].Status.StartTime == nil {
				return failedJobs[j].Status.StartTime != nil
			}
			return failedJobs[i].Status.StartTime.Before(failedJobs[j].Status.StartTime)
		})
		for i, job := range failedJobs {
			if int32(i) >= int32(len(failedJobs))-*cronJob.Spec.FailedJobsHistoryLimit {
				break
			}
			if err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Log.Error(err, "unable to delete old failed job", "job", job)
			} else {
				log.Log.V(0).Info("deleted old failed job", "job", job)
			}
		}
	}

	if cronJob.Spec.SuccessfulJobsHistoryLimit != nil {
		sort.Slice(successfulJobs, func(i, j int) bool {
			if successfulJobs[i].Status.StartTime == nil {
				return successfulJobs[j].Status.StartTime != nil
			}
			return successfulJobs[i].Status.StartTime.Before(successfulJobs[j].Status.StartTime)
		})
		for i, job := range successfulJobs {
			if int32(i) >= int32(len(successfulJobs))-*cronJob.Spec.SuccessfulJobsHistoryLimit {
				break
			}
			if err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); (err) != nil {
				log.Log.Error(err, "unable to delete old successful job", "job", job)
			} else {
				log.Log.V(0).Info("deleted old successful job", "job", job)
			}
		}
	}

	if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
		log.Log.V(1).Info("cronjob suspended, skipping")
		return ctrl.Result{}, nil
	}

	// понять когда нам нужно создать следующие джобы
	missedRun, nextRun, err := getNextSchedule(&cronJob, r.Now())
	if err != nil {
		log.Log.Error(err, "unable to figure out CronJob schedule")
		// нам не надо перезапускать работу пока мы не получим апдейт который
		// пофиксит расписание
		return ctrl.Result{}, nil
	}

	scheduledResult := ctrl.Result{RequeueAfter: nextRun.Sub(r.Now())} // сохранить этот результат, чтобы заюзать позже
	loggr := log.Log.WithValues("now", r.Now(), "next run", nextRun)

	if missedRun.IsZero() {
		loggr.V(1).Info("no upcoming scheduled times, sleeping until next")
		return scheduledResult, nil
	}

	// убедимся что не слишком поздно
	loggr = log.Log.WithValues("current run", missedRun)
	tooLate := false
	if cronJob.Spec.StartingDeadlineSeconds != nil {
		tooLate = missedRun.Add(time.Duration(*cronJob.Spec.StartingDeadlineSeconds) * time.Second).Before(r.Now())
	}
	if tooLate {
		loggr.V(1).Info("missed starting deadline for last run, sleeping till next")
		return scheduledResult, nil
	}
	// понять как запускаем джобу - можно ли запустить конкурентно...
	if cronJob.Spec.ConcurrencyPolicy == batchv1.ForbidConcurrent && len(activeJobs) > 0 {
		loggr.V(1).Info("concurrency policy blocks concurrent runs, skipping", "num active", len(activeJobs))
		return scheduledResult, nil
	}

	// ...или нам надо заменить джобу...
	if cronJob.Spec.ConcurrencyPolicy == batchv1.ReplaceConcurrent {
		for _, activeJob := range activeJobs {
			// нам все равно была уже удалена джоба или нет
			if err := r.Delete(ctx, activeJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Log.Error(err, "unable to delete active job", "job", activeJob)
				return ctrl.Result{}, err
			}
		}
	}

	// создать джобу...
	job, err := constructJobForCronJob(&cronJob, missedRun, r.Scheme)
	if err != nil {
		log.Log.Error(err, "unable to construct job from template")
		// don't bother requeuing until we get a change to the spec
		return scheduledResult, nil
	}

	// ...и добавить ее в кластер
	if err := r.Create(ctx, job); err != nil {
		log.Log.Error(err, "unable to create Job for CronJob", "job", job)
		return ctrl.Result{}, err
	}

	log.Log.V(1).Info("created Job for CronJob run", "job", job)

	// запустимся еще раз чтобы увидеть что работа запустилась и обновить спеку
	return scheduledResult, nil
}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = batchv1.GroupVersion.String()
)

func (r *CronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// ставим реальные часы, у нас тут не тест
	if r.Clock == nil {
		r.Clock = realClock{}
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kbatch.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		// забираем джоб, извлекаем владельца...
		job := rawObj.(*kbatch.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...убедимся что это CronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "CronJob" {
			return nil
		}

		// ...и если так то возвращаем имя
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.CronJob{}).
		Owns(&kbatch.Job{}).
		Complete(r)
}
