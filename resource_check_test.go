package topgun_test

import (
	"bytes"
	"fmt"

	"code.cloudfoundry.org/garden"
	gclient "code.cloudfoundry.org/garden/client"
	gconn "code.cloudfoundry.org/garden/client/connection"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource checking", func() {
	Context("with a tagged worker", func() {
		BeforeEach(func() {
			Deploy("deployments/concourse.yml", "-o", "operations/tagged-worker.yml")
			_ = waitForRunningWorker()
		})

		Context("with tags on the resource", func() {
			BeforeEach(func() {
				By("setting a pipeline that has a tagged resource")
				fly("set-pipeline", "-n", "-c", "pipelines/tagged-resource.yml", "-p", "tagged-resource")

				By("unpausing the pipeline pipeline")
				fly("unpause-pipeline", "-p", "tagged-resource")
			})

			It("places the checking container on the tagged worker", func() {
				By("running the check")
				fly("check-resource", "-r", "tagged-resource/some-resource")

				By("getting the worker name")
				workersTable := flyTable("workers")
				var taggedWorkerName string
				for _, w := range workersTable {
					if w["tags"] == "tagged" {
						taggedWorkerName = w["name"]
					}
				}
				Expect(taggedWorkerName).ToNot(BeEmpty())

				By("checking that the container is on the tagged worker")
				containerTable := flyTable("containers")
				Expect(containerTable).To(HaveLen(1))
				Expect(containerTable[0]["type"]).To(Equal("check"))
				Expect(containerTable[0]["worker"]).To(Equal(taggedWorkerName))
			})
		})
	})

	Context("with multiple workers", func() {
		teamName := "other-team"
		var workerOneGarden garden.Client
		var workerTwoGarden garden.Client

		BeforeEach(func() {
			Deploy("deployments/concourse-different-workers.yml")

			By("setting a pipeline that has a resource")
			fly("set-pipeline", "-n", "-c", "pipelines/resource-check.yml", "-p", "pipeline-1")

			By("unpausing the pipeline pipeline")
			fly("unpause-pipeline", "-p", "pipeline-1")

			By("checking for a version of the resource")
			fly("check-resource", "-r", "pipeline-1/some-resource")

			By("creating another team")
			setTeamSession := spawnFlyInteractive(
				bytes.NewBufferString("y\n"),
				"set-team",
				"--team-name", teamName,
				"--allow-all-users",
			)
			<-setTeamSession.Exited
			Expect(setTeamSession.ExitCode()).To(Equal(0))

			fly("login", "-c", atcExternalURL, "-n", teamName, "-u", atcUsername, "-p", atcPassword)

			By("setting a pipeline that has a resource")
			fly("set-pipeline", "-n", "-c", "pipelines/resource-check.yml", "-p", "pipeline-2")

			By("unpausing the pipeline pipeline")
			fly("unpause-pipeline", "-p", "pipeline-2")

			By("checking for a version of the resource")
			fly("check-resource", "-r", "pipeline-2/some-resource")

			gardens := JobInstances("garden")
			workerOneGarden = gclient.New(gconn.New("tcp", fmt.Sprintf("%s:7777", gardens[0].IP)))
			workerTwoGarden = gclient.New(gconn.New("tcp", fmt.Sprintf("%s:7777", gardens[1].IP)))
		})

		It("only has one check container between workers", func() {
			containers, err := workerOneGarden.Containers(nil)
			Expect(err).ToNot(HaveOccurred())
			otherContainers, err := workerTwoGarden.Containers(nil)
			Expect(err).ToNot(HaveOccurred())

			containers = append(containers, otherContainers...)
			Expect(containers).To(HaveLen(1))
		})
	})
})
