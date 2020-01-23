package v1

import (
	"github.com/riser-platform/riser-server/pkg/util"
	"testing"
	"time"

	"github.com/riser-platform/riser-server/pkg/core"

	"github.com/stretchr/testify/assert"

	"github.com/riser-platform/riser-server/api/v1/model"
)

func Test_mapDeploymentToStatusModel(t *testing.T) {
	deployment := &core.Deployment{
		Name:            "mydeployment",
		StageName:       "mystage",
		RiserGeneration: 4,
		Doc: core.DeploymentDoc{
			Status: &core.DeploymentStatus{
				ObservedRiserGeneration:   3,
				LatestCreatedRevisionName: "rev2",
				LatestReadyRevisionName:   "rev1",
				Revisions: []core.DeploymentRevisionStatus{
					core.DeploymentRevisionStatus{
						Name:                "rev1",
						AvailableReplicas:   1,
						RolloutStatus:       "myrolloutstatus",
						RolloutStatusReason: "myrolloutstatusreason",
						DockerImage:         "mydockerimage",
						RiserGeneration:     3,
						Problems: []core.StatusProblem{
							core.StatusProblem{
								Message: "myproblem1",
								Count:   1,
							},
							core.StatusProblem{
								Message: "myproblem2",
								Count:   2,
							},
						},
					},
					core.DeploymentRevisionStatus{
						Name:                "rev2",
						AvailableReplicas:   1,
						RolloutStatus:       "myrolloutstatus2",
						RolloutStatusReason: "myrolloutstatusreason2",
						DockerImage:         "mydockerimage2",
						RiserGeneration:     4,
					},
				},
				Traffic: []core.DeploymentTrafficStatus{
					core.DeploymentTrafficStatus{
						Percent:      util.PtrInt64(90),
						RevisionName: "rev1",
						Tag:          "r1",
					},
					core.DeploymentTrafficStatus{
						Percent:      util.PtrInt64(10),
						RevisionName: "rev2",
						Tag:          "r2",
					},
				},
			},
		},
	}

	result := mapDeploymentToStatusModel(deployment)

	assert.Equal(t, "mydeployment", result.DeploymentName)
	assert.Equal(t, "mystage", result.StageName)
	assert.Equal(t, int64(3), result.ObservedRiserGeneration)
	assert.Equal(t, int64(4), result.RiserGeneration)
	assert.Equal(t, "rev2", result.LatestCreatedRevisionName)
	assert.Equal(t, "rev1", result.LatestReadyRevisionName)

	// Traffic
	assert.Len(t, result.Traffic, 2)
	assert.Equal(t, "rev1", result.Traffic[0].RevisionName)
	assert.Equal(t, int64(90), *result.Traffic[0].Percent)
	assert.Equal(t, "r1", result.Traffic[0].Tag)
	assert.Equal(t, "rev2", result.Traffic[1].RevisionName)
	assert.Equal(t, int64(10), *result.Traffic[1].Percent)
	assert.Equal(t, "r2", result.Traffic[1].Tag)

	// Revisions
	assert.Len(t, result.Revisions, 2)
	assert.Equal(t, "rev1", result.Revisions[0].Name)
	assert.Equal(t, int32(1), result.Revisions[0].AvailableReplicas)
	assert.Equal(t, "myrolloutstatus", result.Revisions[0].RolloutStatus)
	assert.Equal(t, "myrolloutstatusreason", result.Revisions[0].RolloutStatusReason)
	assert.Equal(t, "mydockerimage", result.Revisions[0].DockerImage)
	assert.Equal(t, int64(3), result.Revisions[0].RiserGeneration)
	assert.Equal(t, "rev2", result.Revisions[1].Name)
	assert.Equal(t, int32(1), result.Revisions[1].AvailableReplicas)
	assert.Equal(t, "myrolloutstatus2", result.Revisions[1].RolloutStatus)
	assert.Equal(t, "myrolloutstatusreason2", result.Revisions[1].RolloutStatusReason)
	assert.Equal(t, "mydockerimage2", result.Revisions[1].DockerImage)
	assert.Equal(t, int64(4), result.Revisions[1].RiserGeneration)

	// Problems
	assert.Len(t, result.Revisions[0].Problems, 2)
	assert.Equal(t, "myproblem1", result.Revisions[0].Problems[0].Message)
	assert.Equal(t, 1, result.Revisions[0].Problems[0].Count)
	assert.Equal(t, "myproblem2", result.Revisions[0].Problems[1].Message)
	assert.Equal(t, 2, result.Revisions[0].Problems[1].Count)
}

func Test_mapDeploymentToStatusModel_NilStatus(t *testing.T) {
	deployment := &core.Deployment{
		Name:      "mydeployment",
		StageName: "mystage",
		Doc:       core.DeploymentDoc{},
	}

	result := mapDeploymentToStatusModel(deployment)

	assert.Equal(t, "mydeployment", result.DeploymentName)
	assert.Equal(t, "mystage", result.StageName)
}

func Test_mapDeploymentStatusFromModel(t *testing.T) {
	deploymentStatus := &model.DeploymentStatusMutable{
		ObservedRiserGeneration:   3,
		LatestReadyRevisionName:   "rev1",
		LatestCreatedRevisionName: "rev2",
		Revisions: []model.DeploymentRevisionStatus{
			model.DeploymentRevisionStatus{
				Name:                "rev1",
				AvailableReplicas:   1,
				RiserGeneration:     2,
				RolloutStatus:       "myrolloutstatus",
				RolloutStatusReason: "myrolloutstatusreason",
				DockerImage:         "mydockerimage",
				Problems: []model.StatusProblem{
					model.StatusProblem{
						Message: "myproblem1",
						Count:   1,
					},
					model.StatusProblem{
						Message: "myproblem2",
						Count:   2,
					},
				},
			},
			model.DeploymentRevisionStatus{
				Name:                "rev2",
				AvailableReplicas:   1,
				RiserGeneration:     3,
				RolloutStatus:       "myrolloutstatus2",
				RolloutStatusReason: "myrolloutstatusreason2",
				DockerImage:         "mydockerimage2",
			},
		},
		Traffic: []model.DeploymentTrafficStatus{
			model.DeploymentTrafficStatus{
				Percent:      util.PtrInt64(90),
				RevisionName: "rev1",
				Tag:          "r1",
			},
			model.DeploymentTrafficStatus{
				Percent:      util.PtrInt64(10),
				RevisionName: "rev2",
				Tag:          "r2",
			},
		},
	}

	now := time.Now().Unix()

	result := mapDeploymentStatusFromModel(deploymentStatus)

	assert.InDelta(t, now, result.LastUpdated.Unix(), 3)
	assert.Equal(t, int64(3), result.ObservedRiserGeneration)
	assert.Equal(t, "rev2", result.LatestCreatedRevisionName)
	assert.Equal(t, "rev1", result.LatestReadyRevisionName)

	// Revisions
	assert.Len(t, result.Revisions, 2)
	assert.Equal(t, int64(2), result.Revisions[0].RiserGeneration)
	assert.Equal(t, int32(1), result.Revisions[0].AvailableReplicas)
	assert.Equal(t, "myrolloutstatus", result.Revisions[0].RolloutStatus)
	assert.Equal(t, "myrolloutstatusreason", result.Revisions[0].RolloutStatusReason)
	assert.Equal(t, "mydockerimage", result.Revisions[0].DockerImage)
	assert.Equal(t, int64(3), result.Revisions[1].RiserGeneration)
	assert.Equal(t, int32(1), result.Revisions[1].AvailableReplicas)
	assert.Equal(t, "myrolloutstatus2", result.Revisions[1].RolloutStatus)
	assert.Equal(t, "myrolloutstatusreason2", result.Revisions[1].RolloutStatusReason)
	assert.Equal(t, "mydockerimage2", result.Revisions[1].DockerImage)

	// Traffic
	assert.Len(t, result.Traffic, 2)
	assert.Equal(t, "rev1", result.Traffic[0].RevisionName)
	assert.Equal(t, int64(90), *result.Traffic[0].Percent)
	assert.Equal(t, "r1", result.Traffic[0].Tag)
	assert.Equal(t, "rev2", result.Traffic[1].RevisionName)
	assert.Equal(t, int64(10), *result.Traffic[1].Percent)
	assert.Equal(t, "r2", result.Traffic[1].Tag)

	// Problems
	assert.Len(t, result.Revisions[0].Problems, 2)
	assert.Equal(t, "myproblem1", result.Revisions[0].Problems[0].Message)
	assert.Equal(t, 1, result.Revisions[0].Problems[0].Count)
	assert.Equal(t, "myproblem2", result.Revisions[0].Problems[1].Message)
	assert.Equal(t, 2, result.Revisions[0].Problems[1].Count)
}
