# Concourse CI Pipelines
This file describes how to publish concourse pipelines

You'll first need to have the fly cli installed, and authenticated

## Main pipeline
This pipeline builds pos-proxy and pushes it to testing bucket on GCS
- To publish the pipeline you'll need to pass the **pipeline** file `pipeline.yml` and the **credentials** file and the **configuration** file `config.yml` that contains the required variables, and you can add `-v key=value` for any custom configurations
- Set pipeline command format:
```
fly -t main sp -p <pipeline-name> -c <pipeline file> -l <path to credentials file> -l <config file>
```
- Example while the current directory is the project root:
```
fly -t main sp -p pos-proxy -c ci/pipeline.yml -l ../credentials.yml -l ci/config.yml
```

## Pipeline to publish to other env:
This pipeline moves a certain version of pos-proxy from testing bucket to a given environment bucket name <target-env-bucket>
- Command format:
```
fly -t main sp -p <pipeline-name> -c <pipeline file> -l <path to credentials file> -v target-env-bucket=<staging|manage|fod> -v build_number=<build number to publish>
```
- Example:
  - To setup the pipeline:
```
fly -t main sp -p pos-proxy-push -c ci/push-pipeline.yml -l ../credentials.yml -v target-env-bucket=staging -v build_number=1234
```

  - To unpause it (In case of first time to setup):
```
fly -t main up -p pos-proxy-push
```

  - To trigger it and watch the logs
```
fly -t main tj -j pos-proxy-push/publish -w
```
