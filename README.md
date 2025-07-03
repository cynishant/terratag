# [<img src="ttlogo.png" width="300" alt="Terratag Logo">](https://terratag.io)

[![ci](https://github.com/cloudyali/terratag/workflows/ci/badge.svg)](https://github.com/cloudyali/terratag/actions?query=workflow%3Aci+branch%3Amaster) [![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fenv0%2Fterratag.svg?type=small)](https://app.fossa.com/projects/git%2Bgithub.com%2Fenv0%2Fterratag?ref=badge_small)

> <sub>Terratag is brought to you with&nbsp;❤️&nbsp; by
> [<img src="logo.svg" width="150">](https://env0.com)
> Let your team manage their own environment in AWS, Azure and Google. <br/>
> Governed by your policies and with complete visibility and cost management.

## What?

Terratag is a CLI tool allowing for tags or labels to be applied across an entire set of OpenTofu/Terraform files, plus comprehensive tag validation and compliance checking. Terratag will apply tags or labels to any AWS, GCP and Azure resources.

## ✨ New: Tag Validation & Compliance Features

🔍 **Advanced Tag Validation** - Validate existing tags against organizational standards
📊 **Multi-Tag Support** - Handle complex scenarios with 12+ tags per resource  
☁️ **Multi-Cloud Analysis** - AWS (1,506 resources), GCP (213 resources), Azure support
📋 **Multiple Report Formats** - Table, JSON, YAML, Markdown outputs
🚀 **CI/CD Integration** - Automated compliance checking for pipelines
🐳 **Docker Support** - Run without local installation using Docker/Docker Compose

### Quick Examples

```bash
# Validate tags against a standard
terratag -validate-only -standard tag-standard.yaml

# Generate compliance report 
terratag -validate-only -standard tag-standard.yaml -report-format markdown -report-output report.md

# CI/CD strict mode
terratag -validate-only -standard tag-standard.yaml -strict-mode -report-format json

# Docker usage (no local installation needed)
docker run --rm -v $(pwd):/workspace terratag:latest -validate-only -standard /standards/tag-standard.yaml

# Docker Compose (recommended)
docker-compose --profile validate up
```

**[📚 Complete Documentation](docs/GETTING_STARTED.md)** | **[🏷️ Tag Validation Guide](docs/TAG_VALIDATION_FEATURES.md)** | **[☁️ AWS Resource Reference](docs/AWS_RESOURCE_TAGGING.md)** | **[🐳 Docker Usage Guide](docs/DOCKER_USAGE.md)**

### Terratag in action

![](https://assets.website-files.com/5dc3f52851595b160ba99670/5f62090d2d532ca35e143133_terratag.gif)

## Why?

Maintaining tags across your application is hard, especially when done manually. Terratag enables you to easily add tags to your existing IaC and benefit from some cross-resource tag applications you wish you had thought of when you had just started writing your OpenTofu/Terraform, saving you tons of time and making future updates easy. [Read more](https://d1.awsstatic.com/whitepapers/aws-tagging-best-practices.pdf) on why tagging is important.

## How?

### Prerequisites

- OpenTofu 1.x or Terraform 0.12 through 1.x.

### Usage

#### Option 1: Docker (Recommended - No Installation Required)

   ```bash
   # Build and run with Docker
   docker build -t terratag:latest .
   docker run --rm -v $(pwd):/workspace terratag:latest -validate-only -standard /standards/tag-standard.yaml
   
   # Or use Docker Compose
   docker-compose --profile validate up
   ```

#### Option 2: Local Installation

1. Install from homebrew:

   ```
   brew install env0/terratag/terratag
   ```

   Or download the latest [release binary](https://github.com/cloudyali/terratag/releases) .

1. Initialize Opentofu/Terraform modules to get provider schema and pull child modules:

   ```bash
    tofu init
   ```

   ```bash
    terraform init
   ```

1. Run Terratag

   ```bash
    terratag -dir=foo/bar -tags={\"environment_id\": \"prod\"}
   ```

   or

   ```bash
    terratag -dir=foo/bar -tags="environment_id=prod,some-tag=value"
   ```

### Example Output

#### Before Terratag

```
|- aws.tf
|- gcp.tf
```

```hcl
# aws.tf
provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags {
    Name        = "My bucket"
  }
}
```

```hcl
#gcp.tf
resource "google_storage_bucket" "static-site" {
  name          = "image-store.com"
  location      = "EU"
  force_destroy = true

  bucket_policy_only = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }
  cors {
    origin          = ["http://image-store.com"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
  labels = {
    "foo" = "bar"
  }
}

```

#### After Terratag

Running `terratag -tags={\"env0_environment_id\":\"dev\",\"env0_project_id\":\"clientA\"}` will output:

```
|- aws.terratag.tf
|- gcp.terratag.tf
|- aws.tf.bak
|- gcp.tf.bak
```

```hcl
# aws.terratag.tf
provider "aws" {
  version = "~> 2.0"
  region  = "us-east-1"
}

resource "aws_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = merge( map("Name", "My bucket" ), local.terratag_added_main)
}
locals {
  terratag_added_main = {"env0_environment_id"="dev","env0_project_id"="clientA"}
}
```

```hcl
# gcp.terratag.tf
resource "google_storage_bucket" "static-site" {
  name          = "image-store.com"
  location      = "EU"
  force_destroy = true

  bucket_policy_only = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }
  cors {
    origin          = ["http://image-store.com"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
  labels = merge( map("foo" , "bar"), local.terratag_added_main)
}
locals {
  terratag_added_main = {"env0_environment_id"="dev","env0_project_id"="clientA"}
}
```

### Optional CLI flags

- `-dir=<path>` - defaults to `.`. Sets the opentofu/terraform folder to tag `.tf` files in
- `-skipTerratagFiles=false` - Dont skip processing `*.terratag.tf` files (when running terratag a second time for the same directory)
- `-rename=false` - Instead of replacing files named `<basename>.tf` with `<basename>.terratag.tf`, keep the original filename
- `-filter=<regular expression>` - defaults to `.*`. Only apply tags to the resource types matched by the regular expression
- `-type=<terraform, terragrunt, or terragrunt-run-all>` - defaults to `terraform` (and `opentofu`). If `terragrunt` is used, tags the files under `.terragrunt-cache` folder. Note: if Terragrunt does not create a `.terragrunt-cache` folder, use the default or omit.
- `-verbose` - Turn on verbose logging
- `-default-to-terraform` By default uses OpenTofu (if installed), if set will use Terraform even when Opentofu is installed
- `--keep-existing-tags` - When set, existing tags will be preserved when merging tags (by default, new tags override existing ones)

Setting options via enviroment variables is also supported. CLI flags have a precedence over envrionment variables.

```
TERRATAG_TAGS
TERRATAG_DIR
TERRATAG_SKIPTERRATAGFILES
TERRATAG_FILTER
TERRATAG_SKIP
TERRATAG_VERBOSE
TERRATAG_RENAME
TERRATAG_TYPE
TERRATAG_DEFAULT_TO_TERRAFORM
TERRATAG_KEEP_EXISTING_TAGS
```

##### See more samples [here](https://github.com/cloudyali/terratag/tree/master/test/fixture)

## Notes

- Resources already having the exact same tag as the one being appended will be overridden
- Supported providers
  - `aws`
  - `google`
  - `azurerm`
  - `azurestack`
  - `azapi`

## Develop

Issues and Pull Requests are very welcome!

### Prerequisites

- Go > 1.23

### Build

```bash
git clone https://github.com/cloudyali/terratag
cd terratag
go mod tidy
go build ./cmd/terratag
```

### Test

#### Structure

The test cases are located under `test/tests`
Each test case placed there should have the following directory structure:

```
my_test
|+ input
  ...            // any depth under /input
     |- main.tf  // this is where we will run all terraform/terratag commands
|- expected
```

- `input` is where you should place the terraform files of your test.
  All commands will be executed wherever down the hierarchy where `main.tf` is located.
  We do that to allow cases where complex nested submodule resolution may take place, and one would like to test how a directory higher up the hierarchy gets resolved.
- `expected` is a directory in which all `.terratag.tf` files will be matched with the output directory

Each terraform version has it's own config file containing the list of test suites to run.
The config file is under `test/fixtures/terraform_xx/config.yaml` where `xx` is the terraform version.

#### What's being tested?

Each test will run:

- `terraform init`
- `terratag`
- `terraform validate`

And finally, will compare the results in `out` with the `expected` directory

#### Running Tests

Tests can only run on a specific Terraform version -

```
go test -run TestTerraformXX
```

We use [tfenv](https://github.com/tfutils/tfenv) to switch between versions. The exact versions used in the CI tests can be found under `test/tfenvconf`.

## Release

1. Create and push a tag locally, in semver format - `git tag v0.1.32 && git push origin --tags`
2. Goto [Github Releases](https://github.com/cloudyali/terratag/releases) and edit the draft created by Release Drafter Bot - it should contain the change log for the release (if not press on Auto-generate release notes). Make sure it's pointing at the tag you created in the previous step and publish the release.
3. Binaries will be automatically generated by the Github action defined in `.github/workflows/release.yml`
4. NPM will automatically pick up on the new version.
