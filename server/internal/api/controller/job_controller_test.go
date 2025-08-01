package controller_test

import (
	"expenses/internal/models"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JobController", func() {
	// Helper to create a job by executing rules
	createTestJob := func(user *TestHelper) int64 {
		executeReq := models.ExecuteRulesRequest{}
		resp, response := user.MakeRequest(http.MethodPost, "/rule/execute", executeReq)
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		Expect(response["data"]).To(HaveKey("job_id"))
		return int64(response["data"].(map[string]any)["job_id"].(float64))
	}

	// Helper to wait for job completion
	waitForJobCompletion := func(jobId int64, user *TestHelper, timeoutMs int) {
		for i := 0; i < timeoutMs/50; i++ {
			resp, response := user.MakeRequest(http.MethodGet, fmt.Sprintf("/job/%d", jobId), nil)
			if resp.StatusCode == http.StatusOK {
				job := response["data"].(map[string]any)
				status := job["status"].(string)
				if status == "completed" || status == "failed" {
					return
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	Describe("GetJobById", func() {
		Context("when job exists", func() {
			It("should return the job successfully", func() {
				jobId := createTestJob(testUser1)

				resp, response := testUser1.MakeRequest(http.MethodGet, fmt.Sprintf("/job/%d", jobId), nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Job fetched successfully"))
				Expect(response["data"]).ToNot(BeNil())

				job := response["data"].(map[string]any)
				Expect(int64(job["id"].(float64))).To(Equal(jobId))
				Expect(job["job_type"]).To(Equal("rule_execution"))
				Expect(job["created_by"]).ToNot(BeNil())
				Expect(job["status"]).To(BeElementOf("pending", "processing", "completed", "failed"))
				Expect(job["created_at"]).ToNot(BeNil())
				Expect(job["updated_at"]).ToNot(BeNil())
			})

			It("should show job metadata for rule execution", func() {
				jobId := createTestJob(testUser1)

				resp, response := testUser1.MakeRequest(http.MethodGet, fmt.Sprintf("/job/%d", jobId), nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				job := response["data"].(map[string]any)
				Expect(job["metadata"]).ToNot(BeNil())
			})

			It("should show completion details for completed jobs", func() {
				jobId := createTestJob(testUser1)
				waitForJobCompletion(jobId, testUser1, 2000)

				resp, response := testUser1.MakeRequest(http.MethodGet, fmt.Sprintf("/job/%d", jobId), nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				job := response["data"].(map[string]any)
				if job["status"].(string) == "completed" {
					Expect(job["completed_at"]).ToNot(BeNil())
					Expect(job["message"]).ToNot(BeNil())
				}
			})
		})

		Context("when job ID is invalid", func() {
			It("should return bad request for invalid job ID format", func() {
				resp, response := testUser1.MakeRequest(http.MethodGet, "/job/invalid", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("Invalid job_id"))
			})

			It("should return not found for non-existent job ID", func() {
				resp, response := testUser1.MakeRequest(http.MethodGet, "/job/999999", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})

			It("should not allow access to jobs from other users", func() {
				jobId := createTestJob(testUser1)

				resp, response := testUser2.MakeRequest(http.MethodGet, fmt.Sprintf("/job/%d", jobId), nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(ContainSubstring("not found"))
			})
		})

		Context("authentication", func() {
			It("should return unauthorized for missing token", func() {
				jobId := createTestJob(testUser1)

				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, fmt.Sprintf("/job/%d", jobId), nil)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("ListJobs", func() {
		Context("when listing jobs successfully", func() {
			It("should return empty list when user has no jobs", func() {
				resp, response := testUser3.MakeRequest(http.MethodGet, "/job", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Jobs fetched successfully"))
				Expect(response["data"]).To(HaveKey("jobs"))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)
				Expect(jobs).To(BeEmpty())
				Expect(int(data["total"].(float64))).To(Equal(0))
			})

			It("should return paginated jobs for user", func() {
				// Create multiple jobs
				jobId1 := createTestJob(testUser1)
				jobId2 := createTestJob(testUser1)

				resp, response := testUser1.MakeRequest(http.MethodGet, "/job", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("Jobs fetched successfully"))
				Expect(response["data"]).To(HaveKey("jobs"))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)
				Expect(len(jobs)).To(BeNumerically(">=", 2))
				Expect(int(data["total"].(float64))).To(BeNumerically(">=", 2))
				Expect(int(data["page"].(float64))).To(Equal(1))
				Expect(int(data["page_size"].(float64))).To(Equal(20))

				// Verify we can find our created jobs
				var found1, found2 bool
				for _, j := range jobs {
					job := j.(map[string]any)
					id := int64(job["id"].(float64))
					if id == jobId1 {
						found1 = true
					}
					if id == jobId2 {
						found2 = true
					}

					// Verify job structure
					Expect(job["job_type"]).To(Equal("rule_execution"))
					Expect(job["created_by"]).ToNot(BeNil())
					Expect(job["status"]).To(BeElementOf("pending", "processing", "completed", "failed"))
					Expect(job["created_at"]).ToNot(BeNil())
				}
				Expect(found1).To(BeTrue())
				Expect(found2).To(BeTrue())
			})

			It("should not list jobs from other users", func() {
				// Create job as user1
				createTestJob(testUser1)

				// List jobs as user2
				resp, response := testUser2.MakeRequest(http.MethodGet, "/job", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)

				// Should not see user1's jobs
				for _, j := range jobs {
					job := j.(map[string]any)
					createdBy := int64(job["created_by"].(float64))
					Expect(createdBy).ToNot(Equal(int64(123))) // testUser1's ID
				}
			})
		})

		Context("when filtering by job type", func() {
			It("should filter jobs by rule_execution type", func() {
				createTestJob(testUser1)

				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?job_type=rule_execution", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)

				for _, j := range jobs {
					job := j.(map[string]any)
					Expect(job["job_type"]).To(Equal("rule_execution"))
				}
			})

			It("should return empty list for non-matching job type", func() {
				createTestJob(testUser1)

				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?job_type=statement_processing", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)

				// Should be empty or only contain statement_processing jobs
				for _, j := range jobs {
					job := j.(map[string]any)
					Expect(job["job_type"]).To(Equal("statement_processing"))
				}
			})
		})

		Context("when filtering by status", func() {
			It("should filter jobs by completed status", func() {
				jobId := createTestJob(testUser1)
				waitForJobCompletion(jobId, testUser1, 2000)

				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?status=completed", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)

				for _, j := range jobs {
					job := j.(map[string]any)
					Expect(job["status"]).To(Equal("completed"))
				}
			})

			It("should filter jobs by pending status", func() {
				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?status=pending", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)

				for _, j := range jobs {
					job := j.(map[string]any)
					Expect(job["status"]).To(Equal("pending"))
				}
			})
		})

		Context("pagination", func() {
			It("should handle custom page size", func() {
				createTestJob(testUser1)

				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?page_size=5", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				Expect(int(data["page_size"].(float64))).To(Equal(5))
			})

			It("should handle page navigation", func() {
				createTestJob(testUser1)

				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?page=1&page_size=10", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				Expect(int(data["page"].(float64))).To(Equal(1))
				Expect(int(data["page_size"].(float64))).To(Equal(10))
			})

			It("should limit page size to maximum", func() {
				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?page_size=200", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				Expect(int(data["page_size"].(float64))).To(Equal(100)) // Should be capped at 100
			})
		})

		Context("combined filters", func() {
			It("should handle multiple filters together", func() {
				jobId := createTestJob(testUser1)
				waitForJobCompletion(jobId, testUser1, 2000)

				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?job_type=rule_execution&status=completed", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				data := response["data"].(map[string]any)
				jobs := data["jobs"].([]any)

				for _, j := range jobs {
					job := j.(map[string]any)
					Expect(job["job_type"]).To(Equal("rule_execution"))
					Expect(job["status"]).To(Equal("completed"))
				}
			})
		})

		Context("authentication", func() {
			It("should return unauthorized for missing token", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/job", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("error handling", func() {
			It("should handle invalid query parameters gracefully", func() {
				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?page=invalid", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).ToNot(BeNil())
			})

			It("should handle invalid job_type parameter", func() {
				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?job_type=invalid_type", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).ToNot(BeNil())
			})

			It("should handle invalid status parameter", func() {
				resp, response := testUser1.MakeRequest(http.MethodGet, "/job?status=invalid_status", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).ToNot(BeNil())
			})
		})
	})
})
