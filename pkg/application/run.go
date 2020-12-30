package application

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/v1alpha2"
	"github.com/oam-dev/kubevela/apis/types"
	"github.com/oam-dev/kubevela/pkg/appfile/storage/driver"
	cmdutil "github.com/oam-dev/kubevela/pkg/commands/util"
	"github.com/oam-dev/kubevela/pkg/oam"
)

// BuildRun will build application and deploy from Appfile
func BuildRun(ctx context.Context, app *driver.Application, client client.Client, env *types.EnvMeta, io cmdutil.IOStreams) error {
	nApp, err := app.InitTasks(io)
	if err != nil {
		return err
	}

	o,scopes, err := nApp.Object(env.Namespace)
	if err != nil {
		return err
	}

	return Run(ctx, client, nil, nil, scopes, o)
}

// Run will deploy OAM objects.
func Run(ctx context.Context, client client.Client,
	ac *v1alpha2.ApplicationConfiguration, comps []*v1alpha2.Component, scopes []oam.Object, app *v1alpha2.Application) error {
	for _, comp := range comps {
		if err := CreateOrUpdateComponent(ctx, client, comp); err != nil {
			return err
		}
	}
	if err := CreateScopes(ctx, client, scopes); err != nil {
		return err
	}
	if ac != nil {
		if err := CreateOrUpdateAppConfig(ctx, client, ac); err != nil {
			return err
		}
	}
	if app != nil {
		if err := CreateOrUpdateApplication(ctx, client, app); err != nil {
			return err
		}
	}
	return nil
}

// CreateOrUpdateComponent will create if not exist and update if exists.
func CreateOrUpdateComponent(ctx context.Context, client client.Client, comp *v1alpha2.Component) error {
	var getc v1alpha2.Component
	key := ctypes.NamespacedName{Name: comp.Name, Namespace: comp.Namespace}
	if err := client.Get(ctx, key, &getc); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return client.Create(ctx, comp)
	}
	comp.ResourceVersion = getc.ResourceVersion
	return client.Update(ctx, comp)
}

// CreateOrUpdateAppConfig will create if not exist and update if exists.
func CreateOrUpdateAppConfig(ctx context.Context, client client.Client, appConfig *v1alpha2.ApplicationConfiguration) error {
	var geta v1alpha2.ApplicationConfiguration
	key := ctypes.NamespacedName{Name: appConfig.Name, Namespace: appConfig.Namespace}
	var exist = true
	if err := client.Get(ctx, key, &geta); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		exist = false
	}
	if !exist {
		return client.Create(ctx, appConfig)
	}
	appConfig.ResourceVersion = geta.ResourceVersion
	return client.Update(ctx, appConfig)
}

// CreateScopes will create all scopes
func CreateScopes(ctx context.Context, client client.Client, scopes []oam.Object) error {
	for _, obj := range scopes {
		key := ctypes.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
		err := client.Get(ctx, key, obj)
		if err == nil {
			return nil
		}
		if !apierrors.IsNotFound(err) {
			return err
		}
		if err = client.Create(ctx, obj); err != nil {
			return err
		}
	}
	return nil
}

// CreateOrUpdateApplication will create if not exist and update if exists.
func CreateOrUpdateApplication(ctx context.Context, client client.Client, app *v1alpha2.Application) error {
	var geta v1alpha2.Application
	key := ctypes.NamespacedName{Name: app.Name, Namespace: app.Namespace}
	var exist = true
	if err := client.Get(ctx, key, &geta); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		exist = false
	}
	if !exist {
		return client.Create(ctx, app)
	}
	app.ResourceVersion = geta.ResourceVersion
	return client.Update(ctx, app)
}
