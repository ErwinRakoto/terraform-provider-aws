package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcrReplicationConfiguration_basic(t *testing.T) {
	resourceName := "aws_ecr_replication_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecr.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrReplicationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrReplicationConfiguration(testAccGetAlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrReplicationConfigurationExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", testAccGetAlternateRegion()),
					testAccCheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrReplicationMultipleRegionConfiguration(testAccGetAlternateRegion(), testAccGetThirdRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrReplicationConfigurationExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", testAccGetAlternateRegion()),
					testAccCheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.1.region", testAccGetThirdRegion()),
					testAccCheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.1.registry_id"),
				),
			},
			{
				Config: testAccAWSEcrReplicationConfiguration(testAccGetAlternateRegion()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrReplicationConfigurationExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", testAccGetAlternateRegion()),
					testAccCheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
				),
			},
			{
				Config: testAccAWSEcrReplicationConfigurationRepositoryFilter(testAccGetAlternateRegion(), "custom-filter1", "custom-filter2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrReplicationConfigurationExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.destination.0.region", testAccGetAlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.0.filter", "custom-filter1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rule.0.repository_filter.1.filter", "custom-filter2"),
					testAccCheckResourceAttrAccountID(resourceName, "replication_configuration.0.rule.0.destination.0.registry_id"),
				),
			},
		},
	})
}

func testAccCheckAWSEcrReplicationConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecrconn
		out, err := conn.DescribeRegistry(&ecr.DescribeRegistryInput{})
		if err != nil {
			return fmt.Errorf("ECR replication rules not found: %w", err)
		}

		if len(out.ReplicationConfiguration.Rules) == 0 {
			return fmt.Errorf("ECR replication rules not found")
		}

		return nil
	}
}

func testAccCheckAWSEcrReplicationConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_replication_configuration" {
			continue
		}

		out, err := conn.DescribeRegistry(&ecr.DescribeRegistryInput{})
		if err != nil {
			return err
		}

		if len(out.ReplicationConfiguration.Rules) != 0 {
			return fmt.Errorf("ECR replication rules found")
		}
	}

	return nil
}

func testAccAWSEcrReplicationConfiguration(region string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
`, region)
}

func testAccAWSEcrReplicationMultipleRegionConfiguration(region1, region2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }


      destination {
        region      = %[2]q
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
`, region1, region2)
}

func testAccAWSEcrReplicationConfigurationRepositoryFilter(region string, filter1Name string, filter2Name string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = %[1]q
        registry_id = data.aws_caller_identity.current.account_id
      }
      repository_filter {
        filter = "%[2]q"
        filter_type = "PREFIX_MATCH"
      }
      repository_filter {
        filter = "%[3]q"
        filter_type = "PREFIX_MATCH"
      }
    }
  }
}
`, region, filter1Name, filter2Name)
}
