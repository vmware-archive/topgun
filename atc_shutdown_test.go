package topgun_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"regexp"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("[#137641079] ATC Shutting down", func() {
	var dbConn *sql.DB

	BeforeEach(func() {
		var err error
		dbConn, err = sql.Open("postgres", fmt.Sprintf("postgres://atc:dummy-password@%s:5432/atc?sslmode=disable", dbIP))
		Expect(err).ToNot(HaveOccurred())
	})

	FContext("with two atcs available", func() {
		BeforeEach(func() {
			Deploy("deployments/two-atcs-one-worker.yml")
		})

		Describe("tracking builds previously tracked by shutdown ATC", func() {

			var (
				stopSession    *gexec.Session
				stopSession2   *gexec.Session
				restartSession *gexec.Session
			)

			BeforeEach(func() {
				stopSession = spawnBosh("stop", "web-2/0")
			})

			AfterEach(func() {
				<-stopSession.Exited
			})

			Context("with a build in-flight", func() {
				var buildSession *gexec.Session
				var buildID string

				BeforeEach(func() {
					buildSession = spawnFly("execute", "-c", "tasks/wait.yml")
					Eventually(buildSession).Should(gbytes.Say("executing build"))

					buildRegex := regexp.MustCompile(`executing build (\d+)`)
					matches := buildRegex.FindSubmatch(buildSession.Out.Contents())
					buildID = string(matches[1])

					Eventually(buildSession).Should(gbytes.Say("waiting for /tmp/stop-waiting"))
				})

				AfterEach(func() {
					buildSession.Signal(os.Interrupt)
					<-buildSession.Exited
				})

				Context("when the other ATC comes back up", func() {

					BeforeEach(func() {
						stopSession2 = spawnBosh("stop", "web/0")
						restartSession = spawnBosh("restart", "web-2/0")
					})

					AfterEach(func() {
						<-stopSession2.Exited
						<-restartSession.Exited
					})

					It("finishes restarting once the build is done", func() {
						By("hijacking the build to tell it to finish")
						Eventually(func() int {
							session := spawnFlyInteractive(
								bytes.NewBufferString("3\n"),
								"hijack",
								"-b", buildID,
								"-s", "one-off",
								"touch", "/tmp/stop-waiting",
							)

							<-session.Exited
							return session.ExitCode()
						}).Should(Equal(0))

						By("waiting for the build to exit")
						Eventually(buildSession).Should(gbytes.Say("done"))
						<-buildSession.Exited
						Expect(buildSession.ExitCode()).To(Equal(0))

					})
				})
			})

		})

	})

})
