## Creating a Devnet

_In the world of Avalanche, we refer to short-lived, test-focused Subnets as devnets._

Using [avalanche-cli](https://github.com/ava-labs/avalanche-cli), we can create a private devnet (running on a
custom Primary Network with traffic scoped to the deployer IP) across any number of regions and nodes
in ~30 minutes with a single script.

### Step 1: Install Dependencies

#### Install and Configure `aws-cli`

Install the [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html). This is used to
authenticate to AWS and manipulate CloudFormation.

Once you've installed the AWS CLI, run `aws configure sso` to login to AWS locally. See
[the avalanche-ops documentation](https://github.com/ava-labs/avalanche-ops#permissions) for additional details.
Set a `profile_name` when logging in, as it will be referenced directly by avalanche-ops commands. **Do not set
an SSO Session Name (not supported).**

#### Install `yq`

Install `yq` using [Homebrew](https://brew.sh/). `yq` is used to manipulate the YAML files produced by
`avalanche-ops`.

You can install `yq` using the following command:

```bash
brew install yq
```

### Step 2: Deploy Devnet on AWS

Once all dependencies are installed, we can create our devnet using a single script. This script deploys
10 validators (equally split between us-west-2, us-east-2, and eu-west-1):

```bash
./scripts/deploy.devnet.sh
```

_When devnet creation is complete, this script will emit commands that can be used to interact
with the devnet (i.e. tx load test) and to tear it down._
