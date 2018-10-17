package topgun_test

import (
	"time"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ATC Rebalance", func() {
	Context("with two atcs available", func() {
		// var atcs []boshInstance
		var atc0Addr string

		BeforeEach(func() {
			By("Configuring two ATCs")
			Deploy("deployments/atc-rebalance.yml", "--scale", "web=2")
			waitForRunningWorker()

			atc0Addr = JobAddressFromHost("web", 0)

			atc0URL := "http://" + atc0Addr
			fly("login", "-c", atc0URL, "-u", atcUsername, "-p", atcPassword)
		})

		Describe("when a rebalance time is configured", func() {
			It("the worker eventually connects to both web nodes over a period of time", func() {
				atc0Hostname := JobContainerId("web", 0)
				atc1Hostname := JobContainerId("web", 1)
				Eventually(func() string {
					workers := flyTable("workers", "-d")
					return workers[0]["garden address"]
				}, time.Second*9, time.Second*1).Should(ContainSubstring(atc0Hostname))
				Eventually(func() string {
					workers := flyTable("workers", "-d")
					return workers[0]["garden address"]
				}, time.Second*9, time.Second*1).Should(ContainSubstring(atc1Hostname))
			})
		})
	})
})
