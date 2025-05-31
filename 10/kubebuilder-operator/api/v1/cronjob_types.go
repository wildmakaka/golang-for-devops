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

package v1

import (
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronJobSpec defines the desired state of CronJob
type CronJobSpec struct {
	//+kubebuilder:validation:MinLength=0

	// Расписание в формате Cron, смотрим здеь https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	//+kubebuilder:validation:Minimum=0

	// Опциональный дедлайн для запуска джобы, если она пропустила свой старт по какой то причине.
	// Пропущенные запуски будут считаться проваленными
	// +optional
	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty"`

	// Указываем как разрулить конкурентное выполнение джобы.
	// Валидные значения:
	// - "Allow" (default): позволять джобам запускаться конкурентно;
	// - "Forbid": запретить конкурентное выполнение, пропустить следующий запуск если текущий не был завершен;
	// - "Replace": отменить текущую джобу и запустить новую вместо нее
	// +optional
	ConcurrencyPolicy ConcurrencyPolicy `json:"concurrencyPolicy,omitempty"`

	// Говорим контроллеру запретить ли последующие исполнения, не применяется на
	// уже запущенные.  Дефолтное значение false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// Определяем работу запускаемую с помощью CronJob.
	JobTemplate batchv1beta1.JobTemplateSpec `json:"jobTemplate"`

	//+kubebuilder:validation:Minimum=0

	// Количество успешно завершенных джобов для очистки.
	// Это указатель, чтобы понять где у нас 0, а где вообще не определена.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	//+kubebuilder:validation:Minimum=0

	// Количество проваленных джобов для очистки.
	// Это указатель, чтобы понять где у нас 0, а где вообще не определена.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`
}

// ConcurrencyPolicy описывает как обращаться с конкурентными джобами.
// Только одна из описанных полиси может быть применена.
// Если ничего не определено то дефолт это
// AllowConcurrent.
// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type ConcurrencyPolicy string

const (
	// AllowConcurrent позволять джобам запускаться конкурентно.
	AllowConcurrent ConcurrencyPolicy = "Allow"

	// ForbidConcurrent запретить конкурентное выполнение, пропустить следующий запуск если текущий не был завершен
	ForbidConcurrent ConcurrencyPolicy = "Forbid"

	// ReplaceConcurrent отменить текущую джобу и запустить новую вместо нее.
	ReplaceConcurrent ConcurrencyPolicy = "Replace"
)

// CronJobStatus определяет статус CronJob
type CronJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Список указателей на текущие джобы.
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// Когда джоба была последний раз запущена.
	// +optional
	LastScheduleTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CronJob is the Schema for the cronjobs API
type CronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSpec   `json:"spec,omitempty"`
	Status CronJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CronJobList contains a list of CronJob
type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJob{}, &CronJobList{})
}
