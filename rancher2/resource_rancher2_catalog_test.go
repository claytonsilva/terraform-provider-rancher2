package rancher2

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	managementClient "github.com/rancher/types/client/management/v3"
)

const (
	testAccRancher2CatalogType   = "rancher2_catalog"
	testAccRancher2CatalogConfig = `
resource "rancher2_catalog" "foo" {
  name = "foo"
  url = "http://foo.com:8080"
  description= "Terraform catalog acceptance test"
}
`

	testAccRancher2CatalogUpdateConfig = `
resource "rancher2_catalog" "foo" {
  name = "foo"
  url = "http://foo.updated.com:8080"
  description= "Terraform catalog acceptance test - updated"
}
 `

	testAccRancher2CatalogRecreateConfig = `
resource "rancher2_catalog" "foo" {
  name = "foo"
  url = "http://foo.com:8080"
  description= "Terraform catalog acceptance test"
}
 `
)

func TestAccRancher2Catalog_basic(t *testing.T) {
	var catalog *managementClient.Catalog

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRancher2CatalogDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRancher2CatalogConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRancher2CatalogExists(testAccRancher2CatalogType+".foo", catalog),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "name", "foo"),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "description", "Terraform catalog acceptance test"),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "url", "http://foo.com:8080"),
				),
			},
			resource.TestStep{
				Config: testAccRancher2CatalogUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRancher2CatalogExists(testAccRancher2CatalogType+".foo", catalog),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "name", "foo"),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "description", "Terraform catalog acceptance test - updated"),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "url", "http://foo.updated.com:8080"),
				),
			},
			resource.TestStep{
				Config: testAccRancher2CatalogRecreateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRancher2CatalogExists(testAccRancher2CatalogType+".foo", catalog),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "name", "foo"),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "description", "Terraform catalog acceptance test"),
					resource.TestCheckResourceAttr(testAccRancher2CatalogType+".foo", "url", "http://foo.com:8080"),
				),
			},
		},
	})
}

func TestAccRancher2Catalog_disappears(t *testing.T) {
	var catalog *managementClient.Catalog

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRancher2CatalogDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRancher2CatalogConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRancher2CatalogExists(testAccRancher2CatalogType+".foo", catalog),
					testAccRancher2CatalogDisappears(catalog),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRancher2CatalogDisappears(cat *managementClient.Catalog) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != testAccRancher2CatalogType {
				continue
			}
			client, err := testAccProvider.Meta().(*Config).ManagementClient()
			if err != nil {
				return err
			}

			cat, err = client.Catalog.ByID(rs.Primary.ID)
			if err != nil {
				if IsNotFound(err) {
					return nil
				}
				return err
			}

			err = client.Catalog.Delete(cat)
			if err != nil {
				return fmt.Errorf("Error removing Catalog: %s", err)
			}

			stateConf := &resource.StateChangeConf{
				Pending:    []string{"active"},
				Target:     []string{"removed"},
				Refresh:    catalogStateRefreshFunc(client, cat.ID),
				Timeout:    10 * time.Minute,
				Delay:      1 * time.Second,
				MinTimeout: 3 * time.Second,
			}

			_, waitErr := stateConf.WaitForState()
			if waitErr != nil {
				return fmt.Errorf(
					"[ERROR] waiting for catalog (%s) to be removed: %s", cat.ID, waitErr)
			}
		}
		return nil

	}
}

func testAccCheckRancher2CatalogExists(n string, cat *managementClient.Catalog) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No catalog ID is set")
		}

		client, err := testAccProvider.Meta().(*Config).ManagementClient()
		if err != nil {
			return err
		}

		foundReg, err := client.Catalog.ByID(rs.Primary.ID)
		if err != nil {
			if IsNotFound(err) {
				return fmt.Errorf("Catalog not found")
			}
			return err
		}

		cat = foundReg

		return nil
	}
}

func testAccCheckRancher2CatalogDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != testAccRancher2CatalogType {
			continue
		}
		client, err := testAccProvider.Meta().(*Config).ManagementClient()
		if err != nil {
			return err
		}

		_, err = client.Catalog.ByID(rs.Primary.ID)
		if err != nil {
			if IsNotFound(err) {
				return nil
			}
			return err
		}
		return fmt.Errorf("Catalog still exists")
	}
	return nil
}
