package v1

import (
	"fmt"
	"net/http"

	"github.com/riser-platform/riser-server/pkg/environment"
	"github.com/riser-platform/riser-server/pkg/git"

	"github.com/riser-platform/riser-server/pkg/state"

	validation "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/labstack/echo/v4"
	"github.com/riser-platform/riser-server/api/v1/model"
	"github.com/riser-platform/riser-server/pkg/core"
	"github.com/riser-platform/riser-server/pkg/rollout"
)

func PutRollout(c echo.Context, rolloutService rollout.Service, environmentService environment.Service, stateRepo git.Repo) error {
	rolloutRequest := &model.RolloutRequest{}

	deploymentName := c.Param("deploymentName")
	namespace := c.Param("namespace")
	envName := c.Param("envName")

	// Validate environment before binding otherwise the client gets a confusing error about route rules when they pass in an invalid environment
	err := environmentService.ValidateDeployable(envName)
	if err != nil {
		return err
	}

	err = c.Bind(rolloutRequest)
	if err != nil {
		return err
	}

	err = validation.Validate(&rolloutRequest)
	if err != nil {
		return core.NewValidationError("Invalid rollout request", err)
	}

	err = rolloutService.UpdateTraffic(core.NewNamespacedName(deploymentName, namespace), envName,
		mapTrafficRulesToDomain(deploymentName, rolloutRequest.Traffic),
		state.NewGitCommitter(stateRepo))
	if err != nil {
		if err == git.ErrNoChanges {
			return c.JSON(http.StatusOK, model.APIResponse{Message: "No changes to rollout"})
		}
		return err
	}
	return nil
}

func mapTrafficRulesToDomain(deploymentName string, traffic []model.TrafficRule) core.TrafficConfig {
	out := core.TrafficConfig{}
	for _, rule := range traffic {
		out = append(out, core.TrafficConfigRule{
			RiserRevision: rule.RiserRevision,
			RevisionName:  fmt.Sprintf("%s-%d", deploymentName, rule.RiserRevision),
			Percent:       rule.Percent,
		})
	}
	return out
}
