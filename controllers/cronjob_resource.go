package controllers

import (
	"context"

	planesv1beta1 "github.com/AlpineCoder/cron-operator/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

//generateCronJob generates the cronjob. Updates on changes to schedule and image
func (r *TransportReconciler) generateCronJob(transport *planesv1beta1.Transport) error {

	planeCronJob := &batchv1beta1.CronJob{}
	newCronJob := &batchv1beta1.CronJob{}
	ctx := context.Background()

	err := r.Get(ctx, types.NamespacedName{Name: transport.Name + "-cronjob", Namespace: transport.Namespace}, planeCronJob)
	if err != nil && errors.IsNotFound(err) {
		if err := r.createCronJob(transport, newCronJob); err != nil {
			return err
		}
		if err = r.Create(ctx, newCronJob); err != nil {
			return err
		}
	} else {
		if planeCronJob.Spec.Schedule != transport.Spec.Schedule {
			klog.Infof("Daily CronJob Schedule differs, %s vs %s", planeCronJob.Spec.Schedule, transport.Spec.Schedule)
			if err := r.createCronJob(transport, newCronJob); err != nil {
				return err
			}
			if err = r.Update(ctx, newCronJob); err != nil {
				return err
			}
		} else if planeCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Command[0] != r.generateCommand(transport) {
			if err := r.createCronJob(transport, newCronJob); err != nil {
				return err
			}
			if err = r.Update(ctx, newCronJob); err != nil {
				return err
			}

		}

	}
	return nil
}

//createCronJob created a Cron Job
func (r *TransportReconciler) createCronJob(transport *planesv1beta1.Transport, newCronJob *batchv1beta1.CronJob) error {
	var (
		historyLimit            int32 = 3
		terminationGracePeriod  int64 = 30
		startingDeadlineSeconds int64 = 3600
	)
	phistorylimit := &historyLimit
	pterminationGracePeriod := &terminationGracePeriod
	pstartingDeadlineSeconds := &startingDeadlineSeconds

	labels := map[string]string{
		"app": transport.Name,
	}
	newCronJob.ObjectMeta = metav1.ObjectMeta{
		Name:        transport.Name + "-cronjob",
		Namespace:   transport.Namespace,
		Labels:      labels,
		Annotations: map[string]string{},
	}

	if transport.Spec.Schedule == "" {
		newCronJob.Spec.Schedule = "5 */1 * * *"
	} else {
		newCronJob.Spec.Schedule = transport.Spec.Schedule
	}

	newCronJob.Spec.ConcurrencyPolicy = "Forbid"
	newCronJob.Spec.SuccessfulJobsHistoryLimit = phistorylimit
	newCronJob.Spec.StartingDeadlineSeconds = pstartingDeadlineSeconds

	newCronJob.Spec.JobTemplate = batchv1beta1.JobTemplateSpec{
		Spec: batchv1.JobSpec{
			BackoffLimit: phistorylimit,

			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "transport-cron",
							Image: transport.Spec.Image,
							Command: []string{
								"/bin/sh",
								"-c",
							},
							ImagePullPolicy: "Always",
							Args: []string{
								transport.Spec.Destination,
							},
						},
					},

					RestartPolicy:                 v1.RestartPolicyNever,
					TerminationGracePeriodSeconds: pterminationGracePeriod,
					DNSPolicy:                     "ClusterFirst",
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "ocpbackupsrv-pull-secret",
						},
					},
				},
			},
		},
	}

	return nil
}

func (r *TransportReconciler) generateCommand(transport *planesv1beta1.Transport) string {
	return "echo \"" + transport.Spec.Destination + "\""
}
