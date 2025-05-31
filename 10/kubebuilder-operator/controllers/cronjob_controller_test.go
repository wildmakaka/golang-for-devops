package controllers

/*
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
// +kubebuilder:docs-gen:collapse=Apache License

/*
Заимпортим все что нам нужно. Также добавим некоторые нужные нам переменные.
*/

import (
	"context"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cronjobv1 "slurm.io/cronjob/api/v1"
)

// +kubebuilder:docs-gen:collapse=Imports
/*
Первый шаг для написания простого интеграционного теста - это создать инстанс CronJob для тестирования.
Обратите внимание что для создания кронджоба, вам надо будет создать структуру кронджоба со всеми спеками.
Когда мы создаем кронджоб, нам нужно создать и все вложенные структуры для него.
Без этих спек, кубернетес не сможет создать и запустить кронджоб.
*/
var _ = Describe("CronJob controller", func() {

	// Определим константы для имен обьектов, таймаутов и интервалов.
	const (
		CronjobName      = "test-cronjob"
		CronjobNamespace = "default"
		JobName          = "test-job"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating CronJob Status", func() {
		It("Should increase CronJob Status.Active count when new Jobs are created", func() {
			By("By creating a new CronJob")
			ctx := context.Background()
			cronJob := &cronjobv1.CronJob{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "batch.tutorial.kubebuilder.io/v1",
					Kind:       "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      CronjobName,
					Namespace: CronjobNamespace,
				},
				Spec: cronjobv1.CronJobSpec{
					Schedule: "1 * * * *",
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							// Для простоты заполним только обязательные поля.
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									// Для простоты заполним только обязательные поля.
									Containers: []v1.Container{
										{
											Name:  "test-container",
											Image: "test-image",
										},
									},
									RestartPolicy: v1.RestartPolicyOnFailure,
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, cronJob)).Should(Succeed())

			/*
				После создания кроджобы, давайте проверим что поля в спеке соотвествуют тому что мы передали.
				Поскольку k8s сервер мог не закончить создания джобы, мы будем использовать Eventually вместо Expect, чтобы дать аписерверу возможность закончить создание обьекта.
				`Eventually()` будет перезапускать функцию раз в interval, до тех пор пока
				(a) вывод функции не совпадет с тем что в `Should()`, или
				(b) количество попыток * interval не превысит timeout.
				Мы определили interval и timeout в наших служебных переменных наверху. Они являются типом Go Duration.
			*/

			cronjobLookupKey := types.NamespacedName{Name: CronjobName, Namespace: CronjobNamespace}
			createdCronjob := &cronjobv1.CronJob{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())
			// Давайте убедимся что наша Schedule была правильно создана
			Expect(createdCronjob.Spec.Schedule).Should(Equal("1 * * * *"))
			/*
				Теперь когда мы создали кронджобу в кластере и все проверили, надо написать тест который протестит само поведение нашего контроллера.
				Давайте протестируем логику нашего CronJob controller, отвечающую за апдейт CronJob.Status.Active
				Мы убедимся что CronJob имеет одну работу и CronJob.Status.Active содержит эту джобу
				Сначала мы должны получить джобу, созданную ранее, и проверить что она не имеет активных джобов.
				Мы заюзаем метод Consistently() для того чтобы убедиться что никаких джобов на появится за определенное время
			*/
			By("By checking the CronJob has zero active Jobs")
			Consistently(func() (int, error) {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return -1, err
				}
				return len(createdCronjob.Status.Active), nil
			}, duration, interval).Should(Equal(0))
			/*
				Далее мы создадим структуру джобы, которая будет принадлежать нашему CronJob
				Мы поставим статус Active в 2, что сэмулировать запуск двух подов, что будет значит что джоба сейчас запущена
				Затем мы берем эту джобы и ставим ей указатель в нашу тестовую CronJob.
				Таким образом мы уверены что наша тестовая джоба принадлежит и, следовательно, будет трекаться нашим тестовым CronJob.
			*/
			By("By creating a new Job")
			testJob := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      JobName,
					Namespace: CronjobNamespace,
				},
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							// Для простоты заполним только обязательные поля.
							Containers: []v1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
				Status: batchv1.JobStatus{
					Active: 2,
				},
			}

			// Обрататие внимание что GroupVersionKind CronJobа требуется чтобы установить овнера.
			kind := reflect.TypeOf(cronjobv1.CronJob{}).Name()
			gvk := cronjobv1.GroupVersion.WithKind(kind)

			controllerRef := metav1.NewControllerRef(createdCronjob, gvk)
			testJob.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})
			Expect(k8sClient.Create(ctx, testJob)).Should(Succeed())
			/*
				Добавление джоба в наш тестовый кронджоб должно стриггерить логику нашего контролера.
				После этого мы можем написать тест который уже проверит именно логику нашего апдейта в Status самого CronJoba
			*/
			By("By checking that the CronJob has one active Job")
			Eventually(func() ([]string, error) {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				if err != nil {
					return nil, err
				}

				names := []string{}
				for _, job := range createdCronjob.Status.Active {
					names = append(names, job.Name)
				}
				return names, nil
			}, timeout, interval).Should(ConsistOf(JobName), "should list our active job %s in the active jobs list in status", JobName)
		})
	})

})

/*
	После этого make test и наслаждаемся жизнью!
*/
