package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/finder"
)

func TestAccAWSServiceCatalogPortfolioShare_basic(t *testing.T) {
	resourceName := "aws_servicecatalog_portfolio_share.test"
	compareName := "aws_servicecatalog_portfolio.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, servicecatalog.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceCatalogPortfolioShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServiceCatalogPortfolioShareConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceCatalogPortfolioShareExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", "en"),
					resource.TestCheckResourceAttr(resourceName, "accepted", "true"),
					resource.TestCheckResourceAttr(resourceName, "principal_id", "arn:aws:organizations::111122223333:organization/o-abcdefghijkl"),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", compareName, "id"),
					resource.TestCheckResourceAttr(resourceName, "share_tag_options", "true"),
					resource.TestCheckResourceAttr(resourceName, "type", servicecatalog.DescribePortfolioShareTypeOrganization),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
					"provisioning_artifact_parameters",
				},
			},
		},
	})
}

func testAccCheckAwsServiceCatalogPortfolioShareDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).scconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_servicecatalog_portfolio_share" {
			continue
		}

		output, err := finder.PortfolioShare(
			conn,
			rs.Primary.Attributes["portfolio_id"],
			rs.Primary.Attributes["share_type"],
			rs.Primary.Attributes["principal_id"],
		)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Portfolio Share (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Service Catalog Portfolio Share (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsServiceCatalogPortfolioShareExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).scconn

		_, err := finder.PortfolioShare(
			conn,
			rs.Primary.Attributes["portfolio_id"],
			rs.Primary.Attributes["share_type"],
			rs.Primary.Attributes["principal_id"],
		)

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return fmt.Errorf("Service Catalog Portfolio Share (%s) not found", rs.Primary.ID)
		}

		if err != nil {
			return fmt.Errorf("error getting Service Catalog Portfolio Share (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSServiceCatalogPortfolioShareConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_organizations_access" "test" {
  enabled = "true"
}

resource "aws_servicecatalog_portfolio" "test" {
  name          = %[1]q
  description   = %[1]q
  provider_name = %[1]q
}

resource "aws_servicecatalog_portfolio_share" "test" {
  accept_language   = "en"
  portfolio_id      = aws_servicecatalog_portfolio.test.id
  share_tag_options = true
  type              = "ORGANIZATION"
  principal_id      = "arn:aws:organizations::111122223333:organization/o-abcdefghijkl"
}
`, rName)
}