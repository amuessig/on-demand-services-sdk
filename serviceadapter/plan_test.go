package serviceadapter_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/on-demand-service-broker-sdk/serviceadapter"
)

var _ = Describe("Plan", func() {
	Describe("(de)serialising", func() {
		planJson := []byte(`{
			  "instance_groups": [
			    {
			      "name": "example-server",
			      "vm_type": "small",
			      "persistent_disk_type": "ten",
			      "networks": [
			        "example-network"
			      ],
			      "azs": [
			        "example-az"
			      ],
			      "instances": 1,
			      "lifecycle": "errand",
						"jobs": [
							{
								"name": "kafka",
								"release": "1.3",
								"properties": {
									"example_bool": true,
									"example_number": 2,
									"example_string": "thing"
								}
							}
						]
			    }
			  ],
			  "properties": {
			    "example": "property"
			   }
			}`)

		expectedPlan := serviceadapter.Plan{
			InstanceGroups: []serviceadapter.InstanceGroup{{
				Name:           "example-server",
				VMType:         "small",
				PersistentDisk: "ten",
				Networks:       []string{"example-network"},
				AZs:            []string{"example-az"},
				Instances:      1,
				Lifecycle:      "errand",
				Jobs: []serviceadapter.Job{{
					Name:    "kafka",
					Release: "1.3",
					Properties: serviceadapter.Properties{
						"example_bool":   true,
						"example_number": 2.0,
						"example_string": "thing",
					},
				}},
			}},
			Properties: serviceadapter.Properties{"example": "property"},
		}

		It("deserialises plan object containing all optional fields from json", func() {
			var plan serviceadapter.Plan
			Expect(json.Unmarshal(planJson, &plan)).To(Succeed())
			Expect(plan).To(Equal(expectedPlan))
		})

		It("serialises plan object containing all optional fields to json", func() {
			Expect(json.Marshal(expectedPlan)).To(MatchJSON(planJson))
		})

		It("serialises plan object containing only mandatory fields to json", func() {
			expectedPlan := serviceadapter.Plan{
				InstanceGroups: []serviceadapter.InstanceGroup{{
					Name:      "example-server",
					VMType:    "small",
					Networks:  []string{"example-network"},
					Instances: 1,
				}},
				Properties: serviceadapter.Properties{},
			}

			planJson := []byte(`{
			  "instance_groups": [
			    {
			      "name": "example-server",
			      "vm_type": "small",
			      "networks": [
			        "example-network"
			      ],
			      "instances": 1
			    }
			  ],
			  "properties": {}
			}`)
			Expect(json.Marshal(expectedPlan)).To(MatchJSON(planJson))
		})
	})

	Describe("validation", func() {
		var plan serviceadapter.Plan

		BeforeEach(func() {
			plan = serviceadapter.Plan{
				InstanceGroups: []serviceadapter.InstanceGroup{{
					Name:      "example-server",
					VMType:    "small",
					Networks:  []string{"example-network"},
					Instances: 1,
				}},
				Properties: serviceadapter.Properties{},
			}
		})

		Context("when nothing is missing", func() {
			It("returns no error", func() {
				Expect(plan.Validate()).ToNot(HaveOccurred())
			})
		})

		Context("when instance groups are missing", func() {
			BeforeEach(func() {
				plan.InstanceGroups = nil
			})

			It("returns an error", func() {
				Expect(plan.Validate()).To(HaveOccurred())
			})
		})

		Context("when vm type is missing", func() {
			BeforeEach(func() {
				plan.InstanceGroups[0].VMType = ""
			})

			It("returns an error", func() {
				Expect(plan.Validate()).To(HaveOccurred())
			})
		})

		Context("when networks is missing", func() {
			BeforeEach(func() {
				plan.InstanceGroups[0].Networks = nil
			})

			It("returns an error", func() {
				Expect(plan.Validate()).To(HaveOccurred())
			})
		})

		Context("when instances is 0", func() {
			BeforeEach(func() {
				plan.InstanceGroups[0].Instances = 0
			})

			It("returns an error", func() {
				Expect(plan.Validate()).To(HaveOccurred())
			})
		})

		Context("when instance group name is missing", func() {
			BeforeEach(func() {
				plan.InstanceGroups[0].Name = ""
			})

			It("returns an error", func() {
				Expect(plan.Validate()).To(HaveOccurred())
			})
		})

		Context("when jobs block is present", func() {
			BeforeEach(func() {
				plan.InstanceGroups[0].Jobs = []serviceadapter.Job{{
					Name:       "job-name",
					Release:    "release-name",
					Properties: serviceadapter.Properties{"example-property": "thing"},
				}}
			})

			Context("when job name is missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].Jobs[0].Name = ""
				})
				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})
			Context("when release is missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].Jobs[0].Release = ""
				})
				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})
			Context("when properties is missing", func() {
				BeforeEach(func() {
					plan.InstanceGroups[0].Jobs[0].Properties = nil
				})
				It("returns an error", func() {
					Expect(plan.Validate()).To(HaveOccurred())
				})
			})
		})

		Context("when multiple jobs are present", func() {
			BeforeEach(func() {
				plan.InstanceGroups[0].Jobs = []serviceadapter.Job{
					{
						Name:       "job-name",
						Release:    "release-name",
						Properties: serviceadapter.Properties{"example-property": "thing"},
					},
					{
						Name: "",
					}}
			})

			It("runs the validation on the second job", func() {
				Expect(plan.Validate()).To(HaveOccurred())
			})
		})

		Context("when multiple instance groups are present", func() {
			BeforeEach(func() {
				plan.InstanceGroups = append(plan.InstanceGroups, serviceadapter.InstanceGroup{
					Name:      "example-server",
					VMType:    "small",
					Networks:  []string{"example-network"},
					Instances: 1,
					Jobs: []serviceadapter.Job{{
						Name: "",
					}},
				})
			})

			It("runs the validation on the second instance group's job", func() {
				Expect(plan.Validate()).To(HaveOccurred())
			})
		})
	})
})
