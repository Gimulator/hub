package client

import (
	"context"
	"reflect"

	hubv1 "github.com/Gimulator/hub/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Client is a wrapper for controller-runtime client
type Client struct {
	client.Client

	Scheme *runtime.Scheme
}

// NewClient returns new instance of Client
func NewClient(c client.Client, scheme *runtime.Scheme) (*Client, error) {
	return &Client{
		Client: c,
		Scheme: scheme,
	}, nil
}

///////////////////////////////////////////////////
////////////////////////////////////////// Room ///
///////////////////////////////////////////////////

// SyncRoom takes a Room object and updates it or creates it if not exists
func (c *Client) SyncRoom(ctx context.Context, room *hubv1.Room) (*hubv1.Room, error) {
	syncedRoom := &hubv1.Room{
		ObjectMeta: metav1.ObjectMeta{
			Name:      room.Name,
			Namespace: room.Namespace,
		},
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff,
		func() error {
			_, err := controllerutil.CreateOrUpdate(ctx, c.Client, syncedRoom, func() error {
				syncedRoom.Status = *room.Status.DeepCopy()
				syncedRoom.Spec = *room.Spec.DeepCopy()
				return nil
			})
			return err
		},
	)

	return syncedRoom, err
}

// GetRoom takes a NamespacedName key and returns a Room object if exists
func (c *Client) GetRoom(ctx context.Context, key types.NamespacedName) (*hubv1.Room, error) {
	room := &hubv1.Room{}

	return room, c.Get(ctx, key, room)
}

// DeleteRoom deletes a Room object
func (c *Client) DeleteRoom(ctx context.Context, room *hubv1.Room) error {
	if err := c.Delete(ctx, room); !errors.IsNotFound(err) {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
////////////////////////////////////////// Pod ///
//////////////////////////////////////////////////

// SyncPod takes a Pod object and updates it or creates it if not exists
func (c *Client) SyncPod(ctx context.Context, pod *corev1.Pod, owner metav1.Object) (*corev1.Pod, error) {
	key := types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}

	syncedPod, err := c.GetPod(ctx, key)
	if errors.IsNotFound(err) {
		syncedPod, err = c.CreatePod(ctx, pod, owner)
		if err != nil {
			return nil, err
		}
		return syncedPod, err
	}
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(pod.Annotations, syncedPod.Annotations) || !reflect.DeepEqual(pod.Labels, syncedPod.Labels) {
		syncedPod.Annotations = pod.DeepCopy().Annotations
		syncedPod.Labels = pod.DeepCopy().Labels

		if owner != nil {
			if err := controllerutil.SetOwnerReference(owner, syncedPod, c.Scheme); err != nil {
				return nil, err
			}
		}

		if err := c.Update(ctx, syncedPod); err != nil {
			return nil, err
		}

		return syncedPod, nil
	}

	return syncedPod, nil
}

// GetPod takes a NamespacedName key and returns a Pod object if exists
func (c *Client) GetPod(ctx context.Context, key types.NamespacedName) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	return pod, c.Get(ctx, key, pod)
}

func (c *Client) CreatePod(ctx context.Context, pod *corev1.Pod, owner metav1.Object) (*corev1.Pod, error) {
	syncedPod := pod.DeepCopy()

	if owner != nil {
		if err := controllerutil.SetOwnerReference(owner, syncedPod, c.Scheme); err != nil {
			return nil, err
		}
	}

	err := c.Create(ctx, syncedPod)
	return syncedPod, err
}

// DeletePod deletes a Pod object
func (c *Client) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	if err := c.Delete(ctx, pod); !errors.IsNotFound(err) {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
////////////////////////////////////////// PVC ///
//////////////////////////////////////////////////

// SyncPVC takes a PVC object and updates it or creates it if not exists
func (c *Client) SyncPVC(ctx context.Context, pvc *corev1.PersistentVolumeClaim, owner metav1.Object) (*corev1.PersistentVolumeClaim, error) {
	key := types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}

	syncedPVC, err := c.GetPVC(ctx, key)
	if err != nil && errors.IsNotFound(err) {
		syncedPVC, err = c.CreatePVC(ctx, pvc, owner)
		return syncedPVC, err
	}
	return syncedPVC, nil
}

// GetPVC takes a NamespacedName key and returns a PVC object if exists
func (c *Client) GetPVC(ctx context.Context, key types.NamespacedName) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}

	return pvc, c.Get(ctx, key, pvc)
}

func (c *Client) CreatePVC(ctx context.Context, pvc *corev1.PersistentVolumeClaim, owner metav1.Object) (*corev1.PersistentVolumeClaim, error) {
	syncedPVC := pvc.DeepCopy()

	if owner != nil {
		if err := controllerutil.SetOwnerReference(owner, syncedPVC, c.Scheme); err != nil {
			return nil, err
		}
	}

	err := c.Create(ctx, syncedPVC)
	return syncedPVC, err
}

// DeletePVC deletes a PVC object
func (c *Client) DeletePVC(ctx context.Context, pvc *corev1.PersistentVolumeClaim) error {
	if err := c.Delete(ctx, pvc); !errors.IsNotFound(err) {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
////////////////////////////////////// Service ///
//////////////////////////////////////////////////

// SyncService takes a Service object and updates it or creates it if not exists
func (c *Client) SyncService(ctx context.Context, service *corev1.Service, owner metav1.Object) (*corev1.Service, error) {
	syncedService := &corev1.Service{}
	err := c.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, syncedService)
	if errors.IsNotFound(err) {
		syncedService, err = c.CreateService(ctx, service, owner)
		return syncedService, err
	}
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(service.Spec.Ports, syncedService.Spec.Ports) || !reflect.DeepEqual(service.Spec.Selector, syncedService.Spec.Selector) {
		if err := c.Delete(ctx, syncedService); err != nil {
			return nil, err
		}

		syncedService, err = c.CreateService(ctx, service, owner)
		return syncedService, nil
	}
	return syncedService, nil
}

func (c *Client) CreateService(ctx context.Context, service *corev1.Service, owner metav1.Object) (*corev1.Service, error) {
	syncedService := service.DeepCopy()

	if owner != nil {
		if err := controllerutil.SetOwnerReference(owner, syncedService, c.Scheme); err != nil {
			return nil, err
		}
	}

	err := c.Create(ctx, syncedService)
	return syncedService, err
}

//////////////////////////////////////////////////
///////////////////////////////////// ConfigMap///
//////////////////////////////////////////////////

func (c *Client) SyncConfigMap(ctx context.Context, cm *corev1.ConfigMap, owner metav1.Object) (*corev1.ConfigMap, error) {
	key := types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}

	syncedCM, err := c.GetConfigMap(ctx, key)
	if errors.IsNotFound(err) {
		syncedCM, err = c.CreateConfigMap(ctx, cm, owner)
		return syncedCM, err
	}
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(cm.Data, syncedCM.Data) {
		syncedCM = cm.DeepCopy()

		if owner != nil {
			if err := controllerutil.SetOwnerReference(owner, syncedCM, c.Scheme); err != nil {
				return nil, err
			}
		}

		if err := c.Update(ctx, syncedCM); err != nil {
			return nil, err
		}
	}
	return syncedCM, nil
}

func (c *Client) GetConfigMap(ctx context.Context, key types.NamespacedName) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}

	return cm, c.Get(ctx, key, cm)
}

func (c *Client) CreateConfigMap(ctx context.Context, cm *corev1.ConfigMap, owner metav1.Object) (*corev1.ConfigMap, error) {
	syncedConfigMap := cm.DeepCopy()

	if owner != nil {
		if err := controllerutil.SetOwnerReference(owner, syncedConfigMap, c.Scheme); err != nil {
			return nil, err
		}
	}

	err := c.Create(ctx, syncedConfigMap)
	return syncedConfigMap, err
}
