package state

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/riser-platform/riser-server/pkg/state/resources"

	"github.com/pkg/errors"
	"github.com/riser-platform/riser-server/pkg/core"
	"github.com/riser-platform/riser-server/pkg/util"
)

type getResourcePathFunc func(resource KubeResource) string

// RenderGeneric is used for generic resources (e.g. riser app namespaces). They will be placed in the root of the namespaced folder.
func RenderGeneric(stage string, resources ...KubeResource) ([]core.ResourceFile, error) {
	return renderKubeResources(func(resource KubeResource) string {
		return getGenericResourcesPath(stage, resource)
	}, resources...)
}

func RenderSealedSecret(app, stage string, sealedSecret *resources.SealedSecret) ([]core.ResourceFile, error) {
	return renderKubeResources(func(resource KubeResource) string {
		return getSecretScmPath(app, stage, sealedSecret)
	}, sealedSecret)
}

func RenderDeployment(deployment *core.Deployment, deploymentResources ...KubeResource) ([]core.ResourceFile, error) {
	files, err := renderKubeResources(func(resource KubeResource) string {
		return getDeploymentScmPath(deployment.DeploymentMeta, resource)
	}, deploymentResources...)

	if err != nil {
		return nil, err
	}

	appConfigFile, err := renderAppConfig(deployment)
	if err != nil {
		return nil, err
	}
	files = append(files, *appConfigFile)

	return files, nil
}

func renderAppConfig(deployment *core.Deployment) (*core.ResourceFile, error) {
	serialized, err := util.ToYaml(deployment.App)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error serializing app config"))
	}
	return &core.ResourceFile{
		Name:     getAppConfigScmPath(deployment),
		Contents: serialized,
	}, nil
}

func renderKubeResources(pathFunc getResourcePathFunc, resources ...KubeResource) ([]core.ResourceFile, error) {
	files := []core.ResourceFile{}
	for _, resource := range resources {
		serialized, err := util.ToYaml(resource)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error serializing resource %q", resource.GetObjectKind()))
		}
		files = append(files, core.ResourceFile{
			Name:     pathFunc(resource),
			Contents: serialized,
		})
	}
	return files, nil
}

func getDeploymentScmPath(deploymentMeta core.DeploymentMeta, resource KubeResource) string {
	return strings.ToLower(filepath.Join(
		getPlatformResourcesPath(deploymentMeta.Stage),
		deploymentMeta.Namespace,
		"deployments",
		deploymentMeta.Name,
		getFileNameFromResource(resource)))
}

func getSecretScmPath(app string, stage string, sealedSecret KubeResource) string {
	return strings.ToLower(filepath.Join(
		getPlatformResourcesPath(stage),
		sealedSecret.GetNamespace(),
		"secrets",
		app,
		getFileNameFromResource(sealedSecret)))
}

func getAppConfigScmPath(deployment *core.Deployment) string {
	return strings.ToLower(filepath.Join(
		"stages",
		deployment.Stage,
		"configs",
		deployment.Namespace,
		deployment.App.Name,
		fmt.Sprintf("%s.yaml", deployment.Name)))
}

func getPlatformResourcesPath(stageName string) string {
	return strings.ToLower(filepath.Join(
		"stages",
		stageName,
		"kube-resources",
		"riser-managed",
	))
}

func getGenericResourcesPath(stageName string, resource KubeResource) string {
	return strings.ToLower(filepath.Join(
		getPlatformResourcesPath(stageName),
		resource.GetNamespace(),
		getFileNameFromResource(resource),
	))
}

func getFileNameFromResource(resource KubeResource) string {
	return strings.ToLower(fmt.Sprintf("%s.%s.yaml", resource.GetObjectKind().GroupVersionKind().Kind, resource.GetName()))
}
