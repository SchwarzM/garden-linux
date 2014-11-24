package uid_pool_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/garden-linux/old/linux_backend/uid_pool"
)

var _ = Describe("Unix UID pool", func() {
	Describe("acquiring", func() {
		It("returns the next available UID block from the pool", func() {
			pool := uid_pool.New(10000, 500, 100)

			uid1, err := pool.Acquire()
			Ω(err).ShouldNot(HaveOccurred())

			uid2, err := pool.Acquire()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(uid1).Should(Equal(uint32(10000)))
			Ω(uid2).Should(Equal(uint32(10100)))
		})

		Context("when the pool is exhausted", func() {
			It("returns an error", func() {
				pool := uid_pool.New(10000, 500, 100)

				for i := 0; i < 5; i++ {
					_, err := pool.Acquire()
					Ω(err).ShouldNot(HaveOccurred())
				}

				_, err := pool.Acquire()
				Ω(err).Should(HaveOccurred())
			})
		})
	})

	Describe("removing", func() {
		It("acquires a specific UID from the pool", func() {
			pool := uid_pool.New(10000, 2, 1)

			err := pool.Remove(10000)
			Ω(err).ShouldNot(HaveOccurred())

			uid, err := pool.Acquire()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(uid).Should(Equal(uint32(10001)))

			_, err = pool.Acquire()
			Ω(err).Should(HaveOccurred())
		})

		Context("when the resource is already acquired", func() {
			It("returns a UIDTakenError", func() {
				pool := uid_pool.New(10000, 2, 1)

				uid, err := pool.Acquire()
				Ω(err).ShouldNot(HaveOccurred())

				err = pool.Remove(uid)
				Ω(err).Should(Equal(uid_pool.UIDTakenError{uid}))
			})
		})
	})

	Describe("releasing", func() {
		It("places a uid back at the end of the pool", func() {
			pool := uid_pool.New(10000, 20, 10)

			uid1, err := pool.Acquire()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(uid1).Should(Equal(uint32(10000)))

			pool.Release(uid1)

			uid2, err := pool.Acquire()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(uid2).Should(Equal(uint32(10010)))

			nextUID, err := pool.Acquire()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(nextUID).Should(Equal(uint32(10000)))
		})

		Context("when the released uid is out of the range", func() {
			It("does not add it to the pool", func() {
				pool := uid_pool.New(10000, 0, 1)

				pool.Release(20000)

				_, err := pool.Acquire()
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when the released uid is not the start of a block", func() {
			It("does not add it to the pool", func() {
				pool := uid_pool.New(100, 5, 6)

				pool.Release(101)

				pool.Acquire()           // -> 100
				_, err := pool.Acquire() // -> 106 = out of range

				Ω(err).Should(HaveOccurred())
			})
		})
	})
})
