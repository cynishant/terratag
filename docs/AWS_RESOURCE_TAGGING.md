# AWS Resource Tagging Support Reference

This document provides a comprehensive overview of AWS resource tagging support as defined in the Terratag project. This information is extracted from the AWS provider schema and indicates which resources support tagging operations.

## Executive Summary

- **Total AWS Resources**: 1,506
- **Taggable Resources**: 736 (48.9%)
- **Non-Taggable Resources**: 770 (51.1%)
- **AWS Services Covered**: 244

## Service Categories

### Fully Taggable Services (70 services)
Services where 100% of resources support tagging:

| Service | Total Resources | All Taggable |
|---------|----------------|--------------|
| datasync | 13 | ✅ |
| fsx | 11 | ✅ |
| imagebuilder | 9 | ✅ |
| dms | 8 | ✅ |
| appmesh | 7 | ✅ |
| finspace | 7 | ✅ |
| memorydb | 7 | ✅ |
| default | 6 | ✅ |
| gamelift | 6 | ✅ |
| workspacesweb | 6 | ✅ |
| batch | 4 | ✅ |
| evidently | 4 | ✅ |
| route53recoveryreadiness | 4 | ✅ |
| transcribe | 4 | ✅ |
| workspaces | 4 | ✅ |
| cleanrooms | 3 | ✅ |
| codepipeline | 3 | ✅ |
| ivs | 3 | ✅ |
| mskconnect | 3 | ✅ |
| pinpointsmsvoicev2 | 3 | ✅ |
| rekognition | 3 | ✅ |
| appintegrations | 2 | ✅ |
| budgets | 2 | ✅ |
| chatbot | 2 | ✅ |
| codeconnections | 2 | ✅ |
| comprehend | 2 | ✅ |
| emrcontainers | 2 | ✅ |
| ivschat | 2 | ✅ |
| keyspaces | 2 | ✅ |
| mq | 2 | ✅ |
| networkmonitor | 2 | ✅ |
| qldb | 2 | ✅ |
| resourceexplorer2 | 2 | ✅ |
| rolesanywhere | 2 | ✅ |
| ssmincidents | 2 | ✅ |
| timestreamwrite | 2 | ✅ |
| applicationinsights | 1 | ✅ |
| bcmdataexports | 1 | ✅ |
| chimesdkmediapipelines | 1 | ✅ |
| codeguruprofiler | 1 | ✅ |
| codegurureviewer | 1 | ✅ |
| codestarnotifications | 1 | ✅ |
| cur | 1 | ✅ |
| customer | 1 | ✅ |
| dlm | 1 | ✅ |
| docdbelastic | 1 | ✅ |
| drs | 1 | ✅ |
| egress | 1 | ✅ |
| emrserverless | 1 | ✅ |
| fis | 1 | ✅ |
| flow | 1 | ✅ |
| instance | 1 | ✅ |
| internetmonitor | 1 | ✅ |
| key | 1 | ✅ |
| mwaa | 1 | ✅ |
| nat | 1 | ✅ |
| neptunegraph | 1 | ✅ |
| notificationscontacts | 1 | ✅ |
| osis | 1 | ✅ |
| pipes | 1 | ✅ |
| placement | 1 | ✅ |
| qbusiness | 1 | ✅ |
| rbin | 1 | ✅ |
| resiliencehub | 1 | ✅ |
| serverlessapplicationrepository | 1 | ✅ |
| ssmquicksetup | 1 | ✅ |
| subnet | 1 | ✅ |
| swf | 1 | ✅ |
| timestreaminfluxdb | 1 | ✅ |
| timestreamquery | 1 | ✅ |

### Partially Taggable Services (148 services)
Services with mixed tagging support:

| Service | Total Resources | Taggable | Non-Taggable | Support % |
|---------|----------------|----------|--------------|-----------|
| ec2 | 48 | 24 | 24 | 50.0% |
| vpc | 39 | 17 | 22 | 43.6% |
| iam | 34 | 9 | 25 | 26.5% |
| cloudwatch | 31 | 13 | 18 | 41.9% |
| sagemaker | 30 | 25 | 5 | 83.3% |
| api | 26 | 8 | 18 | 30.8% |
| route53 | 26 | 8 | 18 | 30.8% |
| s3 | 26 | 4 | 22 | 15.4% |
| lightsail | 23 | 9 | 14 | 39.1% |
| redshift | 23 | 11 | 12 | 47.8% |
| glue | 20 | 11 | 9 | 55.0% |
| quicksight | 20 | 9 | 11 | 45.0% |
| dx | 19 | 8 | 11 | 42.1% |
| iot | 19 | 10 | 9 | 52.6% |
| networkmanager | 19 | 13 | 6 | 68.4% |
| cloudfront | 16 | 2 | 14 | 12.5% |
| connect | 16 | 12 | 4 | 75.0% |
| rds | 15 | 10 | 5 | 66.7% |
| securityhub | 15 | 1 | 14 | 6.7% |
| cognito | 14 | 2 | 12 | 14.3% |
| db | 14 | 10 | 4 | 71.4% |
| s3control | 14 | 5 | 9 | 35.7% |
| vpclattice | 14 | 11 | 3 | 78.6% |
| backup | 13 | 6 | 7 | 46.2% |
| config | 13 | 3 | 10 | 23.1% |
| guardduty | 13 | 5 | 8 | 38.5% |
| lambda | 13 | 3 | 10 | 23.1% |
| servicecatalog | 13 | 3 | 10 | 23.1% |
| wafregional | 13 | 4 | 9 | 30.8% |
| apigatewayv2 | 12 | 4 | 8 | 33.3% |
| pinpoint | 12 | 2 | 10 | 16.7% |
| ssm | 12 | 6 | 6 | 50.0% |
| ssoadmin | 12 | 3 | 9 | 25.0% |
| waf | 12 | 4 | 8 | 33.3% |
| sesv2 | 11 | 4 | 7 | 36.4% |
| appsync | 10 | 1 | 9 | 10.0% |
| datazone | 10 | 1 | 9 | 10.0% |
| elasticache | 10 | 8 | 2 | 80.0% |
| lb | 10 | 5 | 5 | 50.0% |
| storagegateway | 10 | 7 | 3 | 70.0% |
| transfer | 10 | 7 | 3 | 70.0% |
| apprunner | 9 | 6 | 3 | 66.7% |
| dynamodb | 9 | 2 | 7 | 22.2% |
| ecr | 9 | 1 | 8 | 11.1% |
| kms | 9 | 4 | 5 | 44.4% |
| macie2 | 9 | 4 | 5 | 44.4% |
| neptune | 9 | 7 | 2 | 77.8% |
| opensearch | 9 | 1 | 8 | 11.1% |
| appconfig | 8 | 6 | 2 | 75.0% |
| auditmanager | 8 | 3 | 5 | 37.5% |
| bedrockagent | 8 | 4 | 4 | 50.0% |
| directory | 8 | 2 | 6 | 25.0% |
| ebs | 8 | 4 | 4 | 50.0% |
| ecs | 8 | 5 | 3 | 62.5% |
| eks | 8 | 7 | 1 | 87.5% |
| emr | 8 | 2 | 6 | 25.0% |
| msk | 8 | 4 | 4 | 50.0% |
| shield | 8 | 2 | 6 | 25.0% |
| appstream | 7 | 3 | 4 | 42.9% |
| chime | 7 | 1 | 6 | 14.3% |
| docdb | 7 | 5 | 2 | 71.4% |
| globalaccelerator | 7 | 3 | 4 | 42.9% |
| grafana | 7 | 1 | 6 | 14.3% |
| network | 7 | 2 | 5 | 28.6% |
| organizations | 7 | 4 | 3 | 57.1% |
| redshiftserverless | 7 | 2 | 5 | 28.6% |
| wafv2 | 7 | 4 | 3 | 57.1% |
| alb | 6 | 4 | 2 | 66.7% |
| athena | 6 | 3 | 3 | 50.0% |
| bedrock | 6 | 4 | 2 | 66.7% |
| codebuild | 6 | 3 | 3 | 50.0% |
| devicefarm | 6 | 5 | 1 | 83.3% |
| efs | 6 | 2 | 4 | 33.3% |
| kendra | 6 | 5 | 1 | 83.3% |
| kinesis | 6 | 4 | 2 | 66.7% |
| lexv2models | 6 | 1 | 5 | 16.7% |
| location | 6 | 5 | 1 | 83.3% |
| networkfirewall | 6 | 4 | 2 | 66.7% |
| opensearchserverless | 6 | 1 | 5 | 16.7% |
| sns | 6 | 1 | 5 | 16.7% |
| verifiedaccess | 6 | 4 | 2 | 66.7% |
| acmpca | 5 | 1 | 4 | 20.0% |
| amplify | 5 | 2 | 3 | 40.0% |
| appfabric | 5 | 4 | 1 | 80.0% |
| cloudformation | 5 | 2 | 3 | 40.0% |
| detective | 5 | 1 | 4 | 20.0% |
| inspector2 | 5 | 1 | 4 | 20.0% |
| media | 5 | 4 | 1 | 80.0% |
| medialive | 5 | 4 | 1 | 80.0% |
| prometheus | 5 | 3 | 2 | 60.0% |
| ram | 5 | 1 | 4 | 20.0% |
| securitylake | 5 | 2 | 3 | 40.0% |
| service | 5 | 4 | 1 | 80.0% |
| verifiedpermissions | 5 | 1 | 4 | 20.0% |
| vpn | 5 | 2 | 3 | 40.0% |
| ami | 4 | 3 | 1 | 75.0% |
| ce | 4 | 3 | 1 | 75.0% |
| chimesdkvoice | 4 | 2 | 2 | 50.0% |
| codeartifact | 4 | 2 | 2 | 50.0% |
| codecommit | 4 | 1 | 3 | 25.0% |
| dataexchange | 4 | 3 | 1 | 75.0% |
| elastic | 4 | 3 | 1 | 75.0% |
| elasticsearch | 4 | 1 | 3 | 25.0% |
| licensemanager | 4 | 1 | 3 | 25.0% |
| notifications | 4 | 1 | 3 | 25.0% |
| schemas | 4 | 3 | 1 | 75.0% |
| secretsmanager | 4 | 1 | 3 | 25.0% |
| sqs | 4 | 1 | 3 | 25.0% |
| ssmcontacts | 4 | 2 | 2 | 50.0% |
| xray | 4 | 2 | 2 | 50.0% |
| appautoscaling | 3 | 1 | 2 | 33.3% |
| cloudtrail | 3 | 2 | 1 | 66.7% |
| codedeploy | 3 | 2 | 1 | 66.7% |
| dax | 3 | 1 | 2 | 33.3% |
| eip | 3 | 1 | 2 | 33.3% |
| fms | 3 | 2 | 1 | 66.7% |
| inspector | 3 | 2 | 1 | 66.7% |
| m2 | 3 | 2 | 1 | 66.7% |
| oam | 3 | 2 | 1 | 66.7% |
| route | 3 | 1 | 2 | 33.3% |
| route53domains | 3 | 2 | 1 | 66.7% |
| route53profiles | 3 | 2 | 1 | 66.7% |
| servicecatalogappregistry | 3 | 2 | 1 | 66.7% |
| sfn | 3 | 2 | 1 | 66.7% |
| signer | 3 | 1 | 2 | 33.3% |
| spot | 3 | 2 | 1 | 66.7% |
| synthetics | 3 | 2 | 1 | 66.7% |
| accessanalyzer | 2 | 1 | 1 | 50.0% |
| acm | 2 | 1 | 1 | 50.0% |
| appflow | 2 | 1 | 1 | 50.0% |
| cloud9 | 2 | 1 | 1 | 50.0% |
| cloudhsm | 2 | 1 | 1 | 50.0% |
| codestarconnections | 2 | 1 | 1 | 50.0% |
| controltower | 2 | 1 | 1 | 50.0% |
| customerprofiles | 2 | 1 | 1 | 50.0% |
| datapipeline | 2 | 1 | 1 | 50.0% |
| dsql | 2 | 1 | 1 | 50.0% |
| ecrpublic | 2 | 1 | 1 | 50.0% |
| elb | 2 | 1 | 1 | 50.0% |
| glacier | 2 | 1 | 1 | 50.0% |
| internet | 2 | 1 | 1 | 50.0% |
| kinesisanalyticsv2 | 2 | 1 | 1 | 50.0% |
| launch | 2 | 1 | 1 | 50.0% |
| paymentcryptography | 2 | 1 | 1 | 50.0% |
| resourcegroups | 2 | 1 | 1 | 50.0% |
| rum | 2 | 1 | 1 | 50.0% |
| scheduler | 2 | 1 | 1 | 50.0% |
| security | 2 | 1 | 1 | 50.0% |

### Non-Taggable Services (26 services)
Services where no resources support tagging:

| Service | Total Resources | None Taggable |
|---------|----------------|---------------|
| ses | 14 | ❌ |
| autoscaling | 8 | ❌ |
| lakeformation | 8 | ❌ |
| s3tables | 5 | ❌ |
| devopsguru | 4 | ❌ |
| lex | 4 | ❌ |
| route53recoverycontrolconfig | 4 | ❌ |
| account | 3 | ❌ |
| codecatalyst | 3 | ❌ |
| identitystore | 3 | ❌ |
| load | 3 | ❌ |
| servicequotas | 3 | ❌ |
| cloudfrontkeyvaluestore | 2 | ❌ |
| cloudsearch | 2 | ❌ |
| computeoptimizer | 2 | ❌ |
| costoptimizationhub | 2 | ❌ |
| elastictranscoder | 2 | ❌ |
| app | 1 | ❌ |
| autoscalingplans | 1 | ❌ |
| cloudcontrolapi | 1 | ❌ |
| main | 1 | ❌ |
| proxy | 1 | ❌ |
| redshiftdata | 1 | ❌ |
| s3outposts | 1 | ❌ |
| snapshot | 1 | ❌ |
| volume | 1 | ❌ |

## Detailed Service Breakdown

### Major Services Analysis


#### EC2 Service
- **Total Resources**: 48
- **Taggable**: 24 (50.0%)
- **Non-Taggable**: 24
- **Category**: Partially Taggable


#### VPC Service
- **Total Resources**: 39
- **Taggable**: 17 (43.6%)
- **Non-Taggable**: 22
- **Category**: Partially Taggable


#### IAM Service
- **Total Resources**: 34
- **Taggable**: 9 (26.5%)
- **Non-Taggable**: 25
- **Category**: Partially Taggable


#### CLOUDWATCH Service
- **Total Resources**: 31
- **Taggable**: 13 (41.9%)
- **Non-Taggable**: 18
- **Category**: Partially Taggable


#### SAGEMAKER Service
- **Total Resources**: 30
- **Taggable**: 25 (83.3%)
- **Non-Taggable**: 5
- **Category**: Partially Taggable


#### API Service
- **Total Resources**: 26
- **Taggable**: 8 (30.8%)
- **Non-Taggable**: 18
- **Category**: Partially Taggable


#### ROUTE53 Service
- **Total Resources**: 26
- **Taggable**: 8 (30.8%)
- **Non-Taggable**: 18
- **Category**: Partially Taggable


#### S3 Service
- **Total Resources**: 26
- **Taggable**: 4 (15.4%)
- **Non-Taggable**: 22
- **Category**: Partially Taggable


#### LIGHTSAIL Service
- **Total Resources**: 23
- **Taggable**: 9 (39.1%)
- **Non-Taggable**: 14
- **Category**: Partially Taggable


#### REDSHIFT Service
- **Total Resources**: 23
- **Taggable**: 11 (47.8%)
- **Non-Taggable**: 12
- **Category**: Partially Taggable


#### GLUE Service
- **Total Resources**: 20
- **Taggable**: 11 (55.0%)
- **Non-Taggable**: 9
- **Category**: Partially Taggable


#### QUICKSIGHT Service
- **Total Resources**: 20
- **Taggable**: 9 (45.0%)
- **Non-Taggable**: 11
- **Category**: Partially Taggable


#### DX Service
- **Total Resources**: 19
- **Taggable**: 8 (42.1%)
- **Non-Taggable**: 11
- **Category**: Partially Taggable


#### IOT Service
- **Total Resources**: 19
- **Taggable**: 10 (52.6%)
- **Non-Taggable**: 9
- **Category**: Partially Taggable


#### NETWORKMANAGER Service
- **Total Resources**: 19
- **Taggable**: 13 (68.4%)
- **Non-Taggable**: 6
- **Category**: Partially Taggable


#### CLOUDFRONT Service
- **Total Resources**: 16
- **Taggable**: 2 (12.5%)
- **Non-Taggable**: 14
- **Category**: Partially Taggable


#### CONNECT Service
- **Total Resources**: 16
- **Taggable**: 12 (75.0%)
- **Non-Taggable**: 4
- **Category**: Partially Taggable


#### RDS Service
- **Total Resources**: 15
- **Taggable**: 10 (66.7%)
- **Non-Taggable**: 5
- **Category**: Partially Taggable


#### SECURITYHUB Service
- **Total Resources**: 15
- **Taggable**: 1 (6.7%)
- **Non-Taggable**: 14
- **Category**: Partially Taggable


#### COGNITO Service
- **Total Resources**: 14
- **Taggable**: 2 (14.3%)
- **Non-Taggable**: 12
- **Category**: Partially Taggable


## Complete Resource Inventory

### Taggable Resources by Service


#### ACCESSANALYZER
- `aws_accessanalyzer_analyzer` ✅

#### ACM
- `aws_acm_certificate` ✅

#### ACMPCA
- `aws_acmpca_certificate_authority` ✅

#### ALB
- `aws_alb` ✅
- `aws_alb_listener` ✅
- `aws_alb_listener_rule` ✅
- `aws_alb_target_group` ✅

#### AMI
- `aws_ami` ✅
- `aws_ami_copy` ✅
- `aws_ami_from_instance` ✅

#### AMPLIFY
- `aws_amplify_app` ✅
- `aws_amplify_branch` ✅

#### API
- `aws_api_gateway_api_key` ✅
- `aws_api_gateway_client_certificate` ✅
- `aws_api_gateway_domain_name` ✅
- `aws_api_gateway_domain_name_access_association` ✅
- `aws_api_gateway_rest_api` ✅
- `aws_api_gateway_stage` ✅
- `aws_api_gateway_usage_plan` ✅
- `aws_api_gateway_vpc_link` ✅

#### APIGATEWAYV2
- `aws_apigatewayv2_api` ✅
- `aws_apigatewayv2_domain_name` ✅
- `aws_apigatewayv2_stage` ✅
- `aws_apigatewayv2_vpc_link` ✅

#### APPAUTOSCALING
- `aws_appautoscaling_target` ✅

#### APPCONFIG
- `aws_appconfig_application` ✅
- `aws_appconfig_configuration_profile` ✅
- `aws_appconfig_deployment` ✅
- `aws_appconfig_deployment_strategy` ✅
- `aws_appconfig_environment` ✅
- `aws_appconfig_extension` ✅

#### APPFABRIC
- `aws_appfabric_app_authorization` ✅
- `aws_appfabric_app_bundle` ✅
- `aws_appfabric_ingestion` ✅
- `aws_appfabric_ingestion_destination` ✅

#### APPFLOW
- `aws_appflow_flow` ✅

#### APPINTEGRATIONS
- `aws_appintegrations_data_integration` ✅
- `aws_appintegrations_event_integration` ✅

#### APPLICATIONINSIGHTS
- `aws_applicationinsights_application` ✅

#### APPMESH
- `aws_appmesh_gateway_route` ✅
- `aws_appmesh_mesh` ✅
- `aws_appmesh_route` ✅
- `aws_appmesh_virtual_gateway` ✅
- `aws_appmesh_virtual_node` ✅
- `aws_appmesh_virtual_router` ✅
- `aws_appmesh_virtual_service` ✅

#### APPRUNNER
- `aws_apprunner_auto_scaling_configuration_version` ✅
- `aws_apprunner_connection` ✅
- `aws_apprunner_observability_configuration` ✅
- `aws_apprunner_service` ✅
- `aws_apprunner_vpc_connector` ✅
- `aws_apprunner_vpc_ingress_connection` ✅

#### APPSTREAM
- `aws_appstream_fleet` ✅
- `aws_appstream_image_builder` ✅
- `aws_appstream_stack` ✅

#### APPSYNC
- `aws_appsync_graphql_api` ✅

#### ATHENA
- `aws_athena_capacity_reservation` ✅
- `aws_athena_data_catalog` ✅
- `aws_athena_workgroup` ✅

#### AUDITMANAGER
- `aws_auditmanager_assessment` ✅
- `aws_auditmanager_control` ✅
- `aws_auditmanager_framework` ✅

#### BACKUP
- `aws_backup_framework` ✅
- `aws_backup_logically_air_gapped_vault` ✅
- `aws_backup_plan` ✅
- `aws_backup_report_plan` ✅
- `aws_backup_restore_testing_plan` ✅
- `aws_backup_vault` ✅

#### BATCH
- `aws_batch_compute_environment` ✅
- `aws_batch_job_definition` ✅
- `aws_batch_job_queue` ✅
- `aws_batch_scheduling_policy` ✅

#### BCMDATAEXPORTS
- `aws_bcmdataexports_export` ✅

#### BEDROCK
- `aws_bedrock_custom_model` ✅
- `aws_bedrock_guardrail` ✅
- `aws_bedrock_inference_profile` ✅
- `aws_bedrock_provisioned_model_throughput` ✅

#### BEDROCKAGENT
- `aws_bedrockagent_agent` ✅
- `aws_bedrockagent_agent_alias` ✅
- `aws_bedrockagent_knowledge_base` ✅
- `aws_bedrockagent_prompt` ✅

#### BUDGETS
- `aws_budgets_budget` ✅
- `aws_budgets_budget_action` ✅

#### CE
- `aws_ce_anomaly_monitor` ✅
- `aws_ce_anomaly_subscription` ✅
- `aws_ce_cost_category` ✅

#### CHATBOT
- `aws_chatbot_slack_channel_configuration` ✅
- `aws_chatbot_teams_channel_configuration` ✅

#### CHIME
- `aws_chime_voice_connector` ✅

#### CHIMESDKMEDIAPIPELINES
- `aws_chimesdkmediapipelines_media_insights_pipeline_configuration` ✅

#### CHIMESDKVOICE
- `aws_chimesdkvoice_sip_media_application` ✅
- `aws_chimesdkvoice_voice_profile_domain` ✅

#### CLEANROOMS
- `aws_cleanrooms_collaboration` ✅
- `aws_cleanrooms_configured_table` ✅
- `aws_cleanrooms_membership` ✅

#### CLOUD9
- `aws_cloud9_environment_ec2` ✅

#### CLOUDFORMATION
- `aws_cloudformation_stack` ✅
- `aws_cloudformation_stack_set` ✅

#### CLOUDFRONT
- `aws_cloudfront_distribution` ✅
- `aws_cloudfront_vpc_origin` ✅

#### CLOUDHSM
- `aws_cloudhsm_v2_cluster` ✅

#### CLOUDTRAIL
- `aws_cloudtrail` ✅
- `aws_cloudtrail_event_data_store` ✅

#### CLOUDWATCH
- `aws_cloudwatch_composite_alarm` ✅
- `aws_cloudwatch_contributor_insight_rule` ✅
- `aws_cloudwatch_contributor_managed_insight_rule` ✅
- `aws_cloudwatch_event_bus` ✅
- `aws_cloudwatch_event_rule` ✅
- `aws_cloudwatch_log_anomaly_detector` ✅
- `aws_cloudwatch_log_delivery` ✅
- `aws_cloudwatch_log_delivery_destination` ✅
- `aws_cloudwatch_log_delivery_source` ✅
- `aws_cloudwatch_log_destination` ✅
- `aws_cloudwatch_log_group` ✅
- `aws_cloudwatch_metric_alarm` ✅
- `aws_cloudwatch_metric_stream` ✅

#### CODEARTIFACT
- `aws_codeartifact_domain` ✅
- `aws_codeartifact_repository` ✅

#### CODEBUILD
- `aws_codebuild_fleet` ✅
- `aws_codebuild_project` ✅
- `aws_codebuild_report_group` ✅

#### CODECOMMIT
- `aws_codecommit_repository` ✅

#### CODECONNECTIONS
- `aws_codeconnections_connection` ✅
- `aws_codeconnections_host` ✅

#### CODEDEPLOY
- `aws_codedeploy_app` ✅
- `aws_codedeploy_deployment_group` ✅

#### CODEGURUPROFILER
- `aws_codeguruprofiler_profiling_group` ✅

#### CODEGURUREVIEWER
- `aws_codegurureviewer_repository_association` ✅

#### CODEPIPELINE
- `aws_codepipeline` ✅
- `aws_codepipeline_custom_action_type` ✅
- `aws_codepipeline_webhook` ✅

#### CODESTARCONNECTIONS
- `aws_codestarconnections_connection` ✅

#### CODESTARNOTIFICATIONS
- `aws_codestarnotifications_notification_rule` ✅

#### COGNITO
- `aws_cognito_identity_pool` ✅
- `aws_cognito_user_pool` ✅

#### COMPREHEND
- `aws_comprehend_document_classifier` ✅
- `aws_comprehend_entity_recognizer` ✅

#### CONFIG
- `aws_config_aggregate_authorization` ✅
- `aws_config_config_rule` ✅
- `aws_config_configuration_aggregator` ✅

#### CONNECT
- `aws_connect_contact_flow` ✅
- `aws_connect_contact_flow_module` ✅
- `aws_connect_hours_of_operation` ✅
- `aws_connect_instance` ✅
- `aws_connect_phone_number` ✅
- `aws_connect_queue` ✅
- `aws_connect_quick_connect` ✅
- `aws_connect_routing_profile` ✅
- `aws_connect_security_profile` ✅
- `aws_connect_user` ✅
- `aws_connect_user_hierarchy_group` ✅
- `aws_connect_vocabulary` ✅

#### CONTROLTOWER
- `aws_controltower_landing_zone` ✅

#### CUR
- `aws_cur_report_definition` ✅

#### CUSTOMER
- `aws_customer_gateway` ✅

#### CUSTOMERPROFILES
- `aws_customerprofiles_domain` ✅

#### DATAEXCHANGE
- `aws_dataexchange_data_set` ✅
- `aws_dataexchange_revision` ✅
- `aws_dataexchange_revision_assets` ✅

#### DATAPIPELINE
- `aws_datapipeline_pipeline` ✅

#### DATASYNC
- `aws_datasync_agent` ✅
- `aws_datasync_location_azure_blob` ✅
- `aws_datasync_location_efs` ✅
- `aws_datasync_location_fsx_lustre_file_system` ✅
- `aws_datasync_location_fsx_ontap_file_system` ✅
- `aws_datasync_location_fsx_openzfs_file_system` ✅
- `aws_datasync_location_fsx_windows_file_system` ✅
- `aws_datasync_location_hdfs` ✅
- `aws_datasync_location_nfs` ✅
- `aws_datasync_location_object_storage` ✅
- `aws_datasync_location_s3` ✅
- `aws_datasync_location_smb` ✅
- `aws_datasync_task` ✅

#### DATAZONE
- `aws_datazone_domain` ✅

#### DAX
- `aws_dax_cluster` ✅

#### DB
- `aws_db_cluster_snapshot` ✅
- `aws_db_event_subscription` ✅
- `aws_db_instance` ✅
- `aws_db_option_group` ✅
- `aws_db_parameter_group` ✅
- `aws_db_proxy` ✅
- `aws_db_proxy_endpoint` ✅
- `aws_db_snapshot` ✅
- `aws_db_snapshot_copy` ✅
- `aws_db_subnet_group` ✅

#### DEFAULT
- `aws_default_network_acl` ✅
- `aws_default_route_table` ✅
- `aws_default_security_group` ✅
- `aws_default_subnet` ✅
- `aws_default_vpc` ✅
- `aws_default_vpc_dhcp_options` ✅

#### DETECTIVE
- `aws_detective_graph` ✅

#### DEVICEFARM
- `aws_devicefarm_device_pool` ✅
- `aws_devicefarm_instance_profile` ✅
- `aws_devicefarm_network_profile` ✅
- `aws_devicefarm_project` ✅
- `aws_devicefarm_test_grid_project` ✅

#### DIRECTORY
- `aws_directory_service_directory` ✅
- `aws_directory_service_region` ✅

#### DLM
- `aws_dlm_lifecycle_policy` ✅

#### DMS
- `aws_dms_certificate` ✅
- `aws_dms_endpoint` ✅
- `aws_dms_event_subscription` ✅
- `aws_dms_replication_config` ✅
- `aws_dms_replication_instance` ✅
- `aws_dms_replication_subnet_group` ✅
- `aws_dms_replication_task` ✅
- `aws_dms_s3_endpoint` ✅

#### DOCDB
- `aws_docdb_cluster` ✅
- `aws_docdb_cluster_instance` ✅
- `aws_docdb_cluster_parameter_group` ✅
- `aws_docdb_event_subscription` ✅
- `aws_docdb_subnet_group` ✅

#### DOCDBELASTIC
- `aws_docdbelastic_cluster` ✅

#### DRS
- `aws_drs_replication_configuration_template` ✅

#### DSQL
- `aws_dsql_cluster` ✅

#### DX
- `aws_dx_connection` ✅
- `aws_dx_hosted_private_virtual_interface_accepter` ✅
- `aws_dx_hosted_public_virtual_interface_accepter` ✅
- `aws_dx_hosted_transit_virtual_interface_accepter` ✅
- `aws_dx_lag` ✅
- `aws_dx_private_virtual_interface` ✅
- `aws_dx_public_virtual_interface` ✅
- `aws_dx_transit_virtual_interface` ✅

#### DYNAMODB
- `aws_dynamodb_table` ✅
- `aws_dynamodb_table_replica` ✅

#### EBS
- `aws_ebs_snapshot` ✅
- `aws_ebs_snapshot_copy` ✅
- `aws_ebs_snapshot_import` ✅
- `aws_ebs_volume` ✅

#### EC2
- `aws_ec2_capacity_block_reservation` ✅
- `aws_ec2_capacity_reservation` ✅
- `aws_ec2_carrier_gateway` ✅
- `aws_ec2_client_vpn_endpoint` ✅
- `aws_ec2_fleet` ✅
- `aws_ec2_host` ✅
- `aws_ec2_instance_connect_endpoint` ✅
- `aws_ec2_local_gateway_route_table_vpc_association` ✅
- `aws_ec2_managed_prefix_list` ✅
- `aws_ec2_network_insights_analysis` ✅
- `aws_ec2_network_insights_path` ✅
- `aws_ec2_traffic_mirror_filter` ✅
- `aws_ec2_traffic_mirror_session` ✅
- `aws_ec2_traffic_mirror_target` ✅
- `aws_ec2_transit_gateway` ✅
- `aws_ec2_transit_gateway_connect` ✅
- `aws_ec2_transit_gateway_connect_peer` ✅
- `aws_ec2_transit_gateway_multicast_domain` ✅
- `aws_ec2_transit_gateway_peering_attachment` ✅
- `aws_ec2_transit_gateway_peering_attachment_accepter` ✅
- `aws_ec2_transit_gateway_policy_table` ✅
- `aws_ec2_transit_gateway_route_table` ✅
- `aws_ec2_transit_gateway_vpc_attachment` ✅
- `aws_ec2_transit_gateway_vpc_attachment_accepter` ✅

#### ECR
- `aws_ecr_repository` ✅

#### ECRPUBLIC
- `aws_ecrpublic_repository` ✅

#### ECS
- `aws_ecs_capacity_provider` ✅
- `aws_ecs_cluster` ✅
- `aws_ecs_service` ✅
- `aws_ecs_task_definition` ✅
- `aws_ecs_task_set` ✅

#### EFS
- `aws_efs_access_point` ✅
- `aws_efs_file_system` ✅

#### EGRESS
- `aws_egress_only_internet_gateway` ✅

#### EIP
- `aws_eip` ✅

#### EKS
- `aws_eks_access_entry` ✅
- `aws_eks_addon` ✅
- `aws_eks_cluster` ✅
- `aws_eks_fargate_profile` ✅
- `aws_eks_identity_provider_config` ✅
- `aws_eks_node_group` ✅
- `aws_eks_pod_identity_association` ✅

#### ELASTIC
- `aws_elastic_beanstalk_application` ✅
- `aws_elastic_beanstalk_application_version` ✅
- `aws_elastic_beanstalk_environment` ✅

#### ELASTICACHE
- `aws_elasticache_cluster` ✅
- `aws_elasticache_parameter_group` ✅
- `aws_elasticache_replication_group` ✅
- `aws_elasticache_reserved_cache_node` ✅
- `aws_elasticache_serverless_cache` ✅
- `aws_elasticache_subnet_group` ✅
- `aws_elasticache_user` ✅
- `aws_elasticache_user_group` ✅

#### ELASTICSEARCH
- `aws_elasticsearch_domain` ✅

#### ELB
- `aws_elb` ✅

#### EMR
- `aws_emr_cluster` ✅
- `aws_emr_studio` ✅

#### EMRCONTAINERS
- `aws_emrcontainers_job_template` ✅
- `aws_emrcontainers_virtual_cluster` ✅

#### EMRSERVERLESS
- `aws_emrserverless_application` ✅

#### EVIDENTLY
- `aws_evidently_feature` ✅
- `aws_evidently_launch` ✅
- `aws_evidently_project` ✅
- `aws_evidently_segment` ✅

#### FINSPACE
- `aws_finspace_kx_cluster` ✅
- `aws_finspace_kx_database` ✅
- `aws_finspace_kx_dataview` ✅
- `aws_finspace_kx_environment` ✅
- `aws_finspace_kx_scaling_group` ✅
- `aws_finspace_kx_user` ✅
- `aws_finspace_kx_volume` ✅

#### FIS
- `aws_fis_experiment_template` ✅

#### FLOW
- `aws_flow_log` ✅

#### FMS
- `aws_fms_policy` ✅
- `aws_fms_resource_set` ✅

#### FSX
- `aws_fsx_backup` ✅
- `aws_fsx_data_repository_association` ✅
- `aws_fsx_file_cache` ✅
- `aws_fsx_lustre_file_system` ✅
- `aws_fsx_ontap_file_system` ✅
- `aws_fsx_ontap_storage_virtual_machine` ✅
- `aws_fsx_ontap_volume` ✅
- `aws_fsx_openzfs_file_system` ✅
- `aws_fsx_openzfs_snapshot` ✅
- `aws_fsx_openzfs_volume` ✅
- `aws_fsx_windows_file_system` ✅

#### GAMELIFT
- `aws_gamelift_alias` ✅
- `aws_gamelift_build` ✅
- `aws_gamelift_fleet` ✅
- `aws_gamelift_game_server_group` ✅
- `aws_gamelift_game_session_queue` ✅
- `aws_gamelift_script` ✅

#### GLACIER
- `aws_glacier_vault` ✅

#### GLOBALACCELERATOR
- `aws_globalaccelerator_accelerator` ✅
- `aws_globalaccelerator_cross_account_attachment` ✅
- `aws_globalaccelerator_custom_routing_accelerator` ✅

#### GLUE
- `aws_glue_catalog_database` ✅
- `aws_glue_connection` ✅
- `aws_glue_crawler` ✅
- `aws_glue_data_quality_ruleset` ✅
- `aws_glue_dev_endpoint` ✅
- `aws_glue_job` ✅
- `aws_glue_ml_transform` ✅
- `aws_glue_registry` ✅
- `aws_glue_schema` ✅
- `aws_glue_trigger` ✅
- `aws_glue_workflow` ✅

#### GRAFANA
- `aws_grafana_workspace` ✅

#### GUARDDUTY
- `aws_guardduty_detector` ✅
- `aws_guardduty_filter` ✅
- `aws_guardduty_ipset` ✅
- `aws_guardduty_malware_protection_plan` ✅
- `aws_guardduty_threatintelset` ✅

#### IAM
- `aws_iam_instance_profile` ✅
- `aws_iam_openid_connect_provider` ✅
- `aws_iam_policy` ✅
- `aws_iam_role` ✅
- `aws_iam_saml_provider` ✅
- `aws_iam_server_certificate` ✅
- `aws_iam_service_linked_role` ✅
- `aws_iam_user` ✅
- `aws_iam_virtual_mfa_device` ✅

#### IMAGEBUILDER
- `aws_imagebuilder_component` ✅
- `aws_imagebuilder_container_recipe` ✅
- `aws_imagebuilder_distribution_configuration` ✅
- `aws_imagebuilder_image` ✅
- `aws_imagebuilder_image_pipeline` ✅
- `aws_imagebuilder_image_recipe` ✅
- `aws_imagebuilder_infrastructure_configuration` ✅
- `aws_imagebuilder_lifecycle_policy` ✅
- `aws_imagebuilder_workflow` ✅

#### INSPECTOR
- `aws_inspector_assessment_template` ✅
- `aws_inspector_resource_group` ✅

#### INSPECTOR2
- `aws_inspector2_filter` ✅

#### INSTANCE
- `aws_instance` ✅

#### INTERNET
- `aws_internet_gateway` ✅

#### INTERNETMONITOR
- `aws_internetmonitor_monitor` ✅

#### IOT
- `aws_iot_authorizer` ✅
- `aws_iot_billing_group` ✅
- `aws_iot_ca_certificate` ✅
- `aws_iot_domain_configuration` ✅
- `aws_iot_policy` ✅
- `aws_iot_provisioning_template` ✅
- `aws_iot_role_alias` ✅
- `aws_iot_thing_group` ✅
- `aws_iot_thing_type` ✅
- `aws_iot_topic_rule` ✅

#### IVS
- `aws_ivs_channel` ✅
- `aws_ivs_playback_key_pair` ✅
- `aws_ivs_recording_configuration` ✅

#### IVSCHAT
- `aws_ivschat_logging_configuration` ✅
- `aws_ivschat_room` ✅

#### KENDRA
- `aws_kendra_data_source` ✅
- `aws_kendra_faq` ✅
- `aws_kendra_index` ✅
- `aws_kendra_query_suggestions_block_list` ✅
- `aws_kendra_thesaurus` ✅

#### KEY
- `aws_key_pair` ✅

#### KEYSPACES
- `aws_keyspaces_keyspace` ✅
- `aws_keyspaces_table` ✅

#### KINESIS
- `aws_kinesis_analytics_application` ✅
- `aws_kinesis_firehose_delivery_stream` ✅
- `aws_kinesis_stream` ✅
- `aws_kinesis_video_stream` ✅

#### KINESISANALYTICSV2
- `aws_kinesisanalyticsv2_application` ✅

#### KMS
- `aws_kms_external_key` ✅
- `aws_kms_key` ✅
- `aws_kms_replica_external_key` ✅
- `aws_kms_replica_key` ✅

#### LAMBDA
- `aws_lambda_code_signing_config` ✅
- `aws_lambda_event_source_mapping` ✅
- `aws_lambda_function` ✅

#### LAUNCH
- `aws_launch_template` ✅

#### LB
- `aws_lb` ✅
- `aws_lb_listener` ✅
- `aws_lb_listener_rule` ✅
- `aws_lb_target_group` ✅
- `aws_lb_trust_store` ✅

#### LEXV2MODELS
- `aws_lexv2models_bot` ✅

#### LICENSEMANAGER
- `aws_licensemanager_license_configuration` ✅

#### LIGHTSAIL
- `aws_lightsail_bucket` ✅
- `aws_lightsail_certificate` ✅
- `aws_lightsail_container_service` ✅
- `aws_lightsail_database` ✅
- `aws_lightsail_disk` ✅
- `aws_lightsail_distribution` ✅
- `aws_lightsail_instance` ✅
- `aws_lightsail_key_pair` ✅
- `aws_lightsail_lb` ✅

#### LOCATION
- `aws_location_geofence_collection` ✅
- `aws_location_map` ✅
- `aws_location_place_index` ✅
- `aws_location_route_calculator` ✅
- `aws_location_tracker` ✅

#### M2
- `aws_m2_application` ✅
- `aws_m2_environment` ✅

#### MACIE2
- `aws_macie2_classification_job` ✅
- `aws_macie2_custom_data_identifier` ✅
- `aws_macie2_findings_filter` ✅
- `aws_macie2_member` ✅

#### MEDIA
- `aws_media_convert_queue` ✅
- `aws_media_package_channel` ✅
- `aws_media_packagev2_channel_group` ✅
- `aws_media_store_container` ✅

#### MEDIALIVE
- `aws_medialive_channel` ✅
- `aws_medialive_input` ✅
- `aws_medialive_input_security_group` ✅
- `aws_medialive_multiplex` ✅

#### MEMORYDB
- `aws_memorydb_acl` ✅
- `aws_memorydb_cluster` ✅
- `aws_memorydb_multi_region_cluster` ✅
- `aws_memorydb_parameter_group` ✅
- `aws_memorydb_snapshot` ✅
- `aws_memorydb_subnet_group` ✅
- `aws_memorydb_user` ✅

#### MQ
- `aws_mq_broker` ✅
- `aws_mq_configuration` ✅

#### MSK
- `aws_msk_cluster` ✅
- `aws_msk_replicator` ✅
- `aws_msk_serverless_cluster` ✅
- `aws_msk_vpc_connection` ✅

#### MSKCONNECT
- `aws_mskconnect_connector` ✅
- `aws_mskconnect_custom_plugin` ✅
- `aws_mskconnect_worker_configuration` ✅

#### MWAA
- `aws_mwaa_environment` ✅

#### NAT
- `aws_nat_gateway` ✅

#### NEPTUNE
- `aws_neptune_cluster` ✅
- `aws_neptune_cluster_endpoint` ✅
- `aws_neptune_cluster_instance` ✅
- `aws_neptune_cluster_parameter_group` ✅
- `aws_neptune_event_subscription` ✅
- `aws_neptune_parameter_group` ✅
- `aws_neptune_subnet_group` ✅

#### NEPTUNEGRAPH
- `aws_neptunegraph_graph` ✅

#### NETWORK
- `aws_network_acl` ✅
- `aws_network_interface` ✅

#### NETWORKFIREWALL
- `aws_networkfirewall_firewall` ✅
- `aws_networkfirewall_firewall_policy` ✅
- `aws_networkfirewall_rule_group` ✅
- `aws_networkfirewall_tls_inspection_configuration` ✅

#### NETWORKMANAGER
- `aws_networkmanager_connect_attachment` ✅
- `aws_networkmanager_connect_peer` ✅
- `aws_networkmanager_connection` ✅
- `aws_networkmanager_core_network` ✅
- `aws_networkmanager_device` ✅
- `aws_networkmanager_dx_gateway_attachment` ✅
- `aws_networkmanager_global_network` ✅
- `aws_networkmanager_link` ✅
- `aws_networkmanager_site` ✅
- `aws_networkmanager_site_to_site_vpn_attachment` ✅
- `aws_networkmanager_transit_gateway_peering` ✅
- `aws_networkmanager_transit_gateway_route_table_attachment` ✅
- `aws_networkmanager_vpc_attachment` ✅

#### NETWORKMONITOR
- `aws_networkmonitor_monitor` ✅
- `aws_networkmonitor_probe` ✅

#### NOTIFICATIONS
- `aws_notifications_notification_configuration` ✅

#### NOTIFICATIONSCONTACTS
- `aws_notificationscontacts_email_contact` ✅

#### OAM
- `aws_oam_link` ✅
- `aws_oam_sink` ✅

#### OPENSEARCH
- `aws_opensearch_domain` ✅

#### OPENSEARCHSERVERLESS
- `aws_opensearchserverless_collection` ✅

#### ORGANIZATIONS
- `aws_organizations_account` ✅
- `aws_organizations_organizational_unit` ✅
- `aws_organizations_policy` ✅
- `aws_organizations_resource_policy` ✅

#### OSIS
- `aws_osis_pipeline` ✅

#### PAYMENTCRYPTOGRAPHY
- `aws_paymentcryptography_key` ✅

#### PINPOINT
- `aws_pinpoint_app` ✅
- `aws_pinpoint_email_template` ✅

#### PINPOINTSMSVOICEV2
- `aws_pinpointsmsvoicev2_configuration_set` ✅
- `aws_pinpointsmsvoicev2_opt_out_list` ✅
- `aws_pinpointsmsvoicev2_phone_number` ✅

#### PIPES
- `aws_pipes_pipe` ✅

#### PLACEMENT
- `aws_placement_group` ✅

#### PROMETHEUS
- `aws_prometheus_rule_group_namespace` ✅
- `aws_prometheus_scraper` ✅
- `aws_prometheus_workspace` ✅

#### QBUSINESS
- `aws_qbusiness_application` ✅

#### QLDB
- `aws_qldb_ledger` ✅
- `aws_qldb_stream` ✅

#### QUICKSIGHT
- `aws_quicksight_analysis` ✅
- `aws_quicksight_dashboard` ✅
- `aws_quicksight_data_set` ✅
- `aws_quicksight_data_source` ✅
- `aws_quicksight_folder` ✅
- `aws_quicksight_namespace` ✅
- `aws_quicksight_template` ✅
- `aws_quicksight_theme` ✅
- `aws_quicksight_vpc_connection` ✅

#### RAM
- `aws_ram_resource_share` ✅

#### RBIN
- `aws_rbin_rule` ✅

#### RDS
- `aws_rds_cluster` ✅
- `aws_rds_cluster_endpoint` ✅
- `aws_rds_cluster_instance` ✅
- `aws_rds_cluster_parameter_group` ✅
- `aws_rds_cluster_snapshot_copy` ✅
- `aws_rds_custom_db_engine_version` ✅
- `aws_rds_global_cluster` ✅
- `aws_rds_integration` ✅
- `aws_rds_reserved_instance` ✅
- `aws_rds_shard_group` ✅

#### REDSHIFT
- `aws_redshift_cluster` ✅
- `aws_redshift_cluster_snapshot` ✅
- `aws_redshift_event_subscription` ✅
- `aws_redshift_hsm_client_certificate` ✅
- `aws_redshift_hsm_configuration` ✅
- `aws_redshift_integration` ✅
- `aws_redshift_parameter_group` ✅
- `aws_redshift_snapshot_copy_grant` ✅
- `aws_redshift_snapshot_schedule` ✅
- `aws_redshift_subnet_group` ✅
- `aws_redshift_usage_limit` ✅

#### REDSHIFTSERVERLESS
- `aws_redshiftserverless_namespace` ✅
- `aws_redshiftserverless_workgroup` ✅

#### REKOGNITION
- `aws_rekognition_collection` ✅
- `aws_rekognition_project` ✅
- `aws_rekognition_stream_processor` ✅

#### RESILIENCEHUB
- `aws_resiliencehub_resiliency_policy` ✅

#### RESOURCEEXPLORER2
- `aws_resourceexplorer2_index` ✅
- `aws_resourceexplorer2_view` ✅

#### RESOURCEGROUPS
- `aws_resourcegroups_group` ✅

#### ROLESANYWHERE
- `aws_rolesanywhere_profile` ✅
- `aws_rolesanywhere_trust_anchor` ✅

#### ROUTE
- `aws_route_table` ✅

#### ROUTE53
- `aws_route53_health_check` ✅
- `aws_route53_resolver_endpoint` ✅
- `aws_route53_resolver_firewall_domain_list` ✅
- `aws_route53_resolver_firewall_rule_group` ✅
- `aws_route53_resolver_firewall_rule_group_association` ✅
- `aws_route53_resolver_query_log_config` ✅
- `aws_route53_resolver_rule` ✅
- `aws_route53_zone` ✅

#### ROUTE53DOMAINS
- `aws_route53domains_domain` ✅
- `aws_route53domains_registered_domain` ✅

#### ROUTE53PROFILES
- `aws_route53profiles_association` ✅
- `aws_route53profiles_profile` ✅

#### ROUTE53RECOVERYREADINESS
- `aws_route53recoveryreadiness_cell` ✅
- `aws_route53recoveryreadiness_readiness_check` ✅
- `aws_route53recoveryreadiness_recovery_group` ✅
- `aws_route53recoveryreadiness_resource_set` ✅

#### RUM
- `aws_rum_app_monitor` ✅

#### S3
- `aws_s3_bucket` ✅
- `aws_s3_bucket_object` ✅
- `aws_s3_object` ✅
- `aws_s3_object_copy` ✅

#### S3CONTROL
- `aws_s3control_access_grant` ✅
- `aws_s3control_access_grants_instance` ✅
- `aws_s3control_access_grants_location` ✅
- `aws_s3control_bucket` ✅
- `aws_s3control_storage_lens_configuration` ✅

#### SAGEMAKER
- `aws_sagemaker_app` ✅
- `aws_sagemaker_app_image_config` ✅
- `aws_sagemaker_code_repository` ✅
- `aws_sagemaker_data_quality_job_definition` ✅
- `aws_sagemaker_device_fleet` ✅
- `aws_sagemaker_domain` ✅
- `aws_sagemaker_endpoint` ✅
- `aws_sagemaker_endpoint_configuration` ✅
- `aws_sagemaker_feature_group` ✅
- `aws_sagemaker_flow_definition` ✅
- `aws_sagemaker_hub` ✅
- `aws_sagemaker_human_task_ui` ✅
- `aws_sagemaker_image` ✅
- `aws_sagemaker_mlflow_tracking_server` ✅
- `aws_sagemaker_model` ✅
- `aws_sagemaker_model_package_group` ✅
- `aws_sagemaker_monitoring_schedule` ✅
- `aws_sagemaker_notebook_instance` ✅
- `aws_sagemaker_notebook_instance_lifecycle_configuration` ✅
- `aws_sagemaker_pipeline` ✅
- `aws_sagemaker_project` ✅
- `aws_sagemaker_space` ✅
- `aws_sagemaker_studio_lifecycle_config` ✅
- `aws_sagemaker_user_profile` ✅
- `aws_sagemaker_workteam` ✅

#### SCHEDULER
- `aws_scheduler_schedule_group` ✅

#### SCHEMAS
- `aws_schemas_discoverer` ✅
- `aws_schemas_registry` ✅
- `aws_schemas_schema` ✅

#### SECRETSMANAGER
- `aws_secretsmanager_secret` ✅

#### SECURITY
- `aws_security_group` ✅

#### SECURITYHUB
- `aws_securityhub_automation_rule` ✅

#### SECURITYLAKE
- `aws_securitylake_data_lake` ✅
- `aws_securitylake_subscriber` ✅

#### SERVERLESSAPPLICATIONREPOSITORY
- `aws_serverlessapplicationrepository_cloudformation_stack` ✅

#### SERVICE
- `aws_service_discovery_http_namespace` ✅
- `aws_service_discovery_private_dns_namespace` ✅
- `aws_service_discovery_public_dns_namespace` ✅
- `aws_service_discovery_service` ✅

#### SERVICECATALOG
- `aws_servicecatalog_portfolio` ✅
- `aws_servicecatalog_product` ✅
- `aws_servicecatalog_provisioned_product` ✅

#### SERVICECATALOGAPPREGISTRY
- `aws_servicecatalogappregistry_application` ✅
- `aws_servicecatalogappregistry_attribute_group` ✅

#### SESV2
- `aws_sesv2_configuration_set` ✅
- `aws_sesv2_contact_list` ✅
- `aws_sesv2_dedicated_ip_pool` ✅
- `aws_sesv2_email_identity` ✅

#### SFN
- `aws_sfn_activity` ✅
- `aws_sfn_state_machine` ✅

#### SHIELD
- `aws_shield_protection` ✅
- `aws_shield_protection_group` ✅

#### SIGNER
- `aws_signer_signing_profile` ✅

#### SNS
- `aws_sns_topic` ✅

#### SPOT
- `aws_spot_fleet_request` ✅
- `aws_spot_instance_request` ✅

#### SQS
- `aws_sqs_queue` ✅

#### SSM
- `aws_ssm_activation` ✅
- `aws_ssm_association` ✅
- `aws_ssm_document` ✅
- `aws_ssm_maintenance_window` ✅
- `aws_ssm_parameter` ✅
- `aws_ssm_patch_baseline` ✅

#### SSMCONTACTS
- `aws_ssmcontacts_contact` ✅
- `aws_ssmcontacts_rotation` ✅

#### SSMINCIDENTS
- `aws_ssmincidents_replication_set` ✅
- `aws_ssmincidents_response_plan` ✅

#### SSMQUICKSETUP
- `aws_ssmquicksetup_configuration_manager` ✅

#### SSOADMIN
- `aws_ssoadmin_application` ✅
- `aws_ssoadmin_permission_set` ✅
- `aws_ssoadmin_trusted_token_issuer` ✅

#### STORAGEGATEWAY
- `aws_storagegateway_cached_iscsi_volume` ✅
- `aws_storagegateway_file_system_association` ✅
- `aws_storagegateway_gateway` ✅
- `aws_storagegateway_nfs_file_share` ✅
- `aws_storagegateway_smb_file_share` ✅
- `aws_storagegateway_stored_iscsi_volume` ✅
- `aws_storagegateway_tape_pool` ✅

#### SUBNET
- `aws_subnet` ✅

#### SWF
- `aws_swf_domain` ✅

#### SYNTHETICS
- `aws_synthetics_canary` ✅
- `aws_synthetics_group` ✅

#### TIMESTREAMINFLUXDB
- `aws_timestreaminfluxdb_db_instance` ✅

#### TIMESTREAMQUERY
- `aws_timestreamquery_scheduled_query` ✅

#### TIMESTREAMWRITE
- `aws_timestreamwrite_database` ✅
- `aws_timestreamwrite_table` ✅

#### TRANSCRIBE
- `aws_transcribe_language_model` ✅
- `aws_transcribe_medical_vocabulary` ✅
- `aws_transcribe_vocabulary` ✅
- `aws_transcribe_vocabulary_filter` ✅

#### TRANSFER
- `aws_transfer_agreement` ✅
- `aws_transfer_certificate` ✅
- `aws_transfer_connector` ✅
- `aws_transfer_profile` ✅
- `aws_transfer_server` ✅
- `aws_transfer_user` ✅
- `aws_transfer_workflow` ✅

#### VERIFIEDACCESS
- `aws_verifiedaccess_endpoint` ✅
- `aws_verifiedaccess_group` ✅
- `aws_verifiedaccess_instance` ✅
- `aws_verifiedaccess_trust_provider` ✅

#### VERIFIEDPERMISSIONS
- `aws_verifiedpermissions_policy_store` ✅

#### VPC
- `aws_vpc` ✅
- `aws_vpc_block_public_access_exclusion` ✅
- `aws_vpc_dhcp_options` ✅
- `aws_vpc_endpoint` ✅
- `aws_vpc_endpoint_service` ✅
- `aws_vpc_ipam` ✅
- `aws_vpc_ipam_pool` ✅
- `aws_vpc_ipam_resource_discovery` ✅
- `aws_vpc_ipam_resource_discovery_association` ✅
- `aws_vpc_ipam_scope` ✅
- `aws_vpc_peering_connection` ✅
- `aws_vpc_peering_connection_accepter` ✅
- `aws_vpc_route_server` ✅
- `aws_vpc_route_server_endpoint` ✅
- `aws_vpc_route_server_peer` ✅
- `aws_vpc_security_group_egress_rule` ✅
- `aws_vpc_security_group_ingress_rule` ✅

#### VPCLATTICE
- `aws_vpclattice_access_log_subscription` ✅
- `aws_vpclattice_listener` ✅
- `aws_vpclattice_listener_rule` ✅
- `aws_vpclattice_resource_configuration` ✅
- `aws_vpclattice_resource_gateway` ✅
- `aws_vpclattice_service` ✅
- `aws_vpclattice_service_network` ✅
- `aws_vpclattice_service_network_resource_association` ✅
- `aws_vpclattice_service_network_service_association` ✅
- `aws_vpclattice_service_network_vpc_association` ✅
- `aws_vpclattice_target_group` ✅

#### VPN
- `aws_vpn_connection` ✅
- `aws_vpn_gateway` ✅

#### WAF
- `aws_waf_rate_based_rule` ✅
- `aws_waf_rule` ✅
- `aws_waf_rule_group` ✅
- `aws_waf_web_acl` ✅

#### WAFREGIONAL
- `aws_wafregional_rate_based_rule` ✅
- `aws_wafregional_rule` ✅
- `aws_wafregional_rule_group` ✅
- `aws_wafregional_web_acl` ✅

#### WAFV2
- `aws_wafv2_ip_set` ✅
- `aws_wafv2_regex_pattern_set` ✅
- `aws_wafv2_rule_group` ✅
- `aws_wafv2_web_acl` ✅

#### WORKSPACES
- `aws_workspaces_connection_alias` ✅
- `aws_workspaces_directory` ✅
- `aws_workspaces_ip_group` ✅
- `aws_workspaces_workspace` ✅

#### WORKSPACESWEB
- `aws_workspacesweb_browser_settings` ✅
- `aws_workspacesweb_data_protection_settings` ✅
- `aws_workspacesweb_ip_access_settings` ✅
- `aws_workspacesweb_network_settings` ✅
- `aws_workspacesweb_user_access_logging_settings` ✅
- `aws_workspacesweb_user_settings` ✅

#### XRAY
- `aws_xray_group` ✅
- `aws_xray_sampling_rule` ✅

### Non-Taggable Resources by Service


#### ACCESSANALYZER
- `aws_accessanalyzer_archive_rule` ❌

#### ACCOUNT
- `aws_account_alternate_contact` ❌
- `aws_account_primary_contact` ❌
- `aws_account_region` ❌

#### ACM
- `aws_acm_certificate_validation` ❌

#### ACMPCA
- `aws_acmpca_certificate` ❌
- `aws_acmpca_certificate_authority_certificate` ❌
- `aws_acmpca_permission` ❌
- `aws_acmpca_policy` ❌

#### ALB
- `aws_alb_listener_certificate` ❌
- `aws_alb_target_group_attachment` ❌

#### AMI
- `aws_ami_launch_permission` ❌

#### AMPLIFY
- `aws_amplify_backend_environment` ❌
- `aws_amplify_domain_association` ❌
- `aws_amplify_webhook` ❌

#### API
- `aws_api_gateway_account` ❌
- `aws_api_gateway_authorizer` ❌
- `aws_api_gateway_base_path_mapping` ❌
- `aws_api_gateway_deployment` ❌
- `aws_api_gateway_documentation_part` ❌
- `aws_api_gateway_documentation_version` ❌
- `aws_api_gateway_gateway_response` ❌
- `aws_api_gateway_integration` ❌
- `aws_api_gateway_integration_response` ❌
- `aws_api_gateway_method` ❌
- `aws_api_gateway_method_response` ❌
- `aws_api_gateway_method_settings` ❌
- `aws_api_gateway_model` ❌
- `aws_api_gateway_request_validator` ❌
- `aws_api_gateway_resource` ❌
- `aws_api_gateway_rest_api_policy` ❌
- `aws_api_gateway_rest_api_put` ❌
- `aws_api_gateway_usage_plan_key` ❌

#### APIGATEWAYV2
- `aws_apigatewayv2_api_mapping` ❌
- `aws_apigatewayv2_authorizer` ❌
- `aws_apigatewayv2_deployment` ❌
- `aws_apigatewayv2_integration` ❌
- `aws_apigatewayv2_integration_response` ❌
- `aws_apigatewayv2_model` ❌
- `aws_apigatewayv2_route` ❌
- `aws_apigatewayv2_route_response` ❌

#### APP
- `aws_app_cookie_stickiness_policy` ❌

#### APPAUTOSCALING
- `aws_appautoscaling_policy` ❌
- `aws_appautoscaling_scheduled_action` ❌

#### APPCONFIG
- `aws_appconfig_extension_association` ❌
- `aws_appconfig_hosted_configuration_version` ❌

#### APPFABRIC
- `aws_appfabric_app_authorization_connection` ❌

#### APPFLOW
- `aws_appflow_connector_profile` ❌

#### APPRUNNER
- `aws_apprunner_custom_domain_association` ❌
- `aws_apprunner_default_auto_scaling_configuration_version` ❌
- `aws_apprunner_deployment` ❌

#### APPSTREAM
- `aws_appstream_directory_config` ❌
- `aws_appstream_fleet_stack_association` ❌
- `aws_appstream_user` ❌
- `aws_appstream_user_stack_association` ❌

#### APPSYNC
- `aws_appsync_api_cache` ❌
- `aws_appsync_api_key` ❌
- `aws_appsync_datasource` ❌
- `aws_appsync_domain_name` ❌
- `aws_appsync_domain_name_api_association` ❌
- `aws_appsync_function` ❌
- `aws_appsync_resolver` ❌
- `aws_appsync_source_api_association` ❌
- `aws_appsync_type` ❌

#### ATHENA
- `aws_athena_database` ❌
- `aws_athena_named_query` ❌
- `aws_athena_prepared_statement` ❌

#### AUDITMANAGER
- `aws_auditmanager_account_registration` ❌
- `aws_auditmanager_assessment_delegation` ❌
- `aws_auditmanager_assessment_report` ❌
- `aws_auditmanager_framework_share` ❌
- `aws_auditmanager_organization_admin_account_registration` ❌

#### AUTOSCALING
- `aws_autoscaling_attachment` ❌
- `aws_autoscaling_group` ❌
- `aws_autoscaling_group_tag` ❌
- `aws_autoscaling_lifecycle_hook` ❌
- `aws_autoscaling_notification` ❌
- `aws_autoscaling_policy` ❌
- `aws_autoscaling_schedule` ❌
- `aws_autoscaling_traffic_source_attachment` ❌

#### AUTOSCALINGPLANS
- `aws_autoscalingplans_scaling_plan` ❌

#### BACKUP
- `aws_backup_global_settings` ❌
- `aws_backup_region_settings` ❌
- `aws_backup_restore_testing_selection` ❌
- `aws_backup_selection` ❌
- `aws_backup_vault_lock_configuration` ❌
- `aws_backup_vault_notifications` ❌
- `aws_backup_vault_policy` ❌

#### BEDROCK
- `aws_bedrock_guardrail_version` ❌
- `aws_bedrock_model_invocation_logging_configuration` ❌

#### BEDROCKAGENT
- `aws_bedrockagent_agent_action_group` ❌
- `aws_bedrockagent_agent_collaborator` ❌
- `aws_bedrockagent_agent_knowledge_base_association` ❌
- `aws_bedrockagent_data_source` ❌

#### CE
- `aws_ce_cost_allocation_tag` ❌

#### CHIME
- `aws_chime_voice_connector_group` ❌
- `aws_chime_voice_connector_logging` ❌
- `aws_chime_voice_connector_origination` ❌
- `aws_chime_voice_connector_streaming` ❌
- `aws_chime_voice_connector_termination` ❌
- `aws_chime_voice_connector_termination_credentials` ❌

#### CHIMESDKVOICE
- `aws_chimesdkvoice_global_settings` ❌
- `aws_chimesdkvoice_sip_rule` ❌

#### CLOUD9
- `aws_cloud9_environment_membership` ❌

#### CLOUDCONTROLAPI
- `aws_cloudcontrolapi_resource` ❌

#### CLOUDFORMATION
- `aws_cloudformation_stack_instances` ❌
- `aws_cloudformation_stack_set_instance` ❌
- `aws_cloudformation_type` ❌

#### CLOUDFRONT
- `aws_cloudfront_cache_policy` ❌
- `aws_cloudfront_continuous_deployment_policy` ❌
- `aws_cloudfront_field_level_encryption_config` ❌
- `aws_cloudfront_field_level_encryption_profile` ❌
- `aws_cloudfront_function` ❌
- `aws_cloudfront_key_group` ❌
- `aws_cloudfront_key_value_store` ❌
- `aws_cloudfront_monitoring_subscription` ❌
- `aws_cloudfront_origin_access_control` ❌
- `aws_cloudfront_origin_access_identity` ❌
- `aws_cloudfront_origin_request_policy` ❌
- `aws_cloudfront_public_key` ❌
- `aws_cloudfront_realtime_log_config` ❌
- `aws_cloudfront_response_headers_policy` ❌

#### CLOUDFRONTKEYVALUESTORE
- `aws_cloudfrontkeyvaluestore_key` ❌
- `aws_cloudfrontkeyvaluestore_keys_exclusive` ❌

#### CLOUDHSM
- `aws_cloudhsm_v2_hsm` ❌

#### CLOUDSEARCH
- `aws_cloudsearch_domain` ❌
- `aws_cloudsearch_domain_service_access_policy` ❌

#### CLOUDTRAIL
- `aws_cloudtrail_organization_delegated_admin_account` ❌

#### CLOUDWATCH
- `aws_cloudwatch_dashboard` ❌
- `aws_cloudwatch_event_api_destination` ❌
- `aws_cloudwatch_event_archive` ❌
- `aws_cloudwatch_event_bus_policy` ❌
- `aws_cloudwatch_event_connection` ❌
- `aws_cloudwatch_event_endpoint` ❌
- `aws_cloudwatch_event_permission` ❌
- `aws_cloudwatch_event_target` ❌
- `aws_cloudwatch_log_account_policy` ❌
- `aws_cloudwatch_log_data_protection_policy` ❌
- `aws_cloudwatch_log_delivery_destination_policy` ❌
- `aws_cloudwatch_log_destination_policy` ❌
- `aws_cloudwatch_log_index_policy` ❌
- `aws_cloudwatch_log_metric_filter` ❌
- `aws_cloudwatch_log_resource_policy` ❌
- `aws_cloudwatch_log_stream` ❌
- `aws_cloudwatch_log_subscription_filter` ❌
- `aws_cloudwatch_query_definition` ❌

#### CODEARTIFACT
- `aws_codeartifact_domain_permissions_policy` ❌
- `aws_codeartifact_repository_permissions_policy` ❌

#### CODEBUILD
- `aws_codebuild_resource_policy` ❌
- `aws_codebuild_source_credential` ❌
- `aws_codebuild_webhook` ❌

#### CODECATALYST
- `aws_codecatalyst_dev_environment` ❌
- `aws_codecatalyst_project` ❌
- `aws_codecatalyst_source_repository` ❌

#### CODECOMMIT
- `aws_codecommit_approval_rule_template` ❌
- `aws_codecommit_approval_rule_template_association` ❌
- `aws_codecommit_trigger` ❌

#### CODEDEPLOY
- `aws_codedeploy_deployment_config` ❌

#### CODESTARCONNECTIONS
- `aws_codestarconnections_host` ❌

#### COGNITO
- `aws_cognito_identity_pool_provider_principal_tag` ❌
- `aws_cognito_identity_pool_roles_attachment` ❌
- `aws_cognito_identity_provider` ❌
- `aws_cognito_managed_user_pool_client` ❌
- `aws_cognito_resource_server` ❌
- `aws_cognito_risk_configuration` ❌
- `aws_cognito_user` ❌
- `aws_cognito_user_group` ❌
- `aws_cognito_user_in_group` ❌
- `aws_cognito_user_pool_client` ❌
- `aws_cognito_user_pool_domain` ❌
- `aws_cognito_user_pool_ui_customization` ❌

#### COMPUTEOPTIMIZER
- `aws_computeoptimizer_enrollment_status` ❌
- `aws_computeoptimizer_recommendation_preferences` ❌

#### CONFIG
- `aws_config_configuration_recorder` ❌
- `aws_config_configuration_recorder_status` ❌
- `aws_config_conformance_pack` ❌
- `aws_config_delivery_channel` ❌
- `aws_config_organization_conformance_pack` ❌
- `aws_config_organization_custom_policy_rule` ❌
- `aws_config_organization_custom_rule` ❌
- `aws_config_organization_managed_rule` ❌
- `aws_config_remediation_configuration` ❌
- `aws_config_retention_configuration` ❌

#### CONNECT
- `aws_connect_bot_association` ❌
- `aws_connect_instance_storage_config` ❌
- `aws_connect_lambda_function_association` ❌
- `aws_connect_user_hierarchy_structure` ❌

#### CONTROLTOWER
- `aws_controltower_control` ❌

#### COSTOPTIMIZATIONHUB
- `aws_costoptimizationhub_enrollment_status` ❌
- `aws_costoptimizationhub_preferences` ❌

#### CUSTOMERPROFILES
- `aws_customerprofiles_profile` ❌

#### DATAEXCHANGE
- `aws_dataexchange_event_action` ❌

#### DATAPIPELINE
- `aws_datapipeline_pipeline_definition` ❌

#### DATAZONE
- `aws_datazone_asset_type` ❌
- `aws_datazone_environment` ❌
- `aws_datazone_environment_blueprint_configuration` ❌
- `aws_datazone_environment_profile` ❌
- `aws_datazone_form_type` ❌
- `aws_datazone_glossary` ❌
- `aws_datazone_glossary_term` ❌
- `aws_datazone_project` ❌
- `aws_datazone_user_profile` ❌

#### DAX
- `aws_dax_parameter_group` ❌
- `aws_dax_subnet_group` ❌

#### DB
- `aws_db_instance_automated_backups_replication` ❌
- `aws_db_instance_role_association` ❌
- `aws_db_proxy_default_target_group` ❌
- `aws_db_proxy_target` ❌

#### DETECTIVE
- `aws_detective_invitation_accepter` ❌
- `aws_detective_member` ❌
- `aws_detective_organization_admin_account` ❌
- `aws_detective_organization_configuration` ❌

#### DEVICEFARM
- `aws_devicefarm_upload` ❌

#### DEVOPSGURU
- `aws_devopsguru_event_sources_config` ❌
- `aws_devopsguru_notification_channel` ❌
- `aws_devopsguru_resource_collection` ❌
- `aws_devopsguru_service_integration` ❌

#### DIRECTORY
- `aws_directory_service_conditional_forwarder` ❌
- `aws_directory_service_log_subscription` ❌
- `aws_directory_service_radius_settings` ❌
- `aws_directory_service_shared_directory` ❌
- `aws_directory_service_shared_directory_accepter` ❌
- `aws_directory_service_trust` ❌

#### DOCDB
- `aws_docdb_cluster_snapshot` ❌
- `aws_docdb_global_cluster` ❌

#### DSQL
- `aws_dsql_cluster_peering` ❌

#### DX
- `aws_dx_bgp_peer` ❌
- `aws_dx_connection_association` ❌
- `aws_dx_connection_confirmation` ❌
- `aws_dx_gateway` ❌
- `aws_dx_gateway_association` ❌
- `aws_dx_gateway_association_proposal` ❌
- `aws_dx_hosted_connection` ❌
- `aws_dx_hosted_private_virtual_interface` ❌
- `aws_dx_hosted_public_virtual_interface` ❌
- `aws_dx_hosted_transit_virtual_interface` ❌
- `aws_dx_macsec_key_association` ❌

#### DYNAMODB
- `aws_dynamodb_contributor_insights` ❌
- `aws_dynamodb_global_table` ❌
- `aws_dynamodb_kinesis_streaming_destination` ❌
- `aws_dynamodb_resource_policy` ❌
- `aws_dynamodb_table_export` ❌
- `aws_dynamodb_table_item` ❌
- `aws_dynamodb_tag` ❌

#### EBS
- `aws_ebs_default_kms_key` ❌
- `aws_ebs_encryption_by_default` ❌
- `aws_ebs_fast_snapshot_restore` ❌
- `aws_ebs_snapshot_block_public_access` ❌

#### EC2
- `aws_ec2_availability_zone_group` ❌
- `aws_ec2_client_vpn_authorization_rule` ❌
- `aws_ec2_client_vpn_network_association` ❌
- `aws_ec2_client_vpn_route` ❌
- `aws_ec2_default_credit_specification` ❌
- `aws_ec2_image_block_public_access` ❌
- `aws_ec2_instance_metadata_defaults` ❌
- `aws_ec2_instance_state` ❌
- `aws_ec2_local_gateway_route` ❌
- `aws_ec2_managed_prefix_list_entry` ❌
- `aws_ec2_serial_console_access` ❌
- `aws_ec2_subnet_cidr_reservation` ❌
- `aws_ec2_tag` ❌
- `aws_ec2_traffic_mirror_filter_rule` ❌
- `aws_ec2_transit_gateway_default_route_table_association` ❌
- `aws_ec2_transit_gateway_default_route_table_propagation` ❌
- `aws_ec2_transit_gateway_multicast_domain_association` ❌
- `aws_ec2_transit_gateway_multicast_group_member` ❌
- `aws_ec2_transit_gateway_multicast_group_source` ❌
- `aws_ec2_transit_gateway_policy_table_association` ❌
- `aws_ec2_transit_gateway_prefix_list_reference` ❌
- `aws_ec2_transit_gateway_route` ❌
- `aws_ec2_transit_gateway_route_table_association` ❌
- `aws_ec2_transit_gateway_route_table_propagation` ❌

#### ECR
- `aws_ecr_account_setting` ❌
- `aws_ecr_lifecycle_policy` ❌
- `aws_ecr_pull_through_cache_rule` ❌
- `aws_ecr_registry_policy` ❌
- `aws_ecr_registry_scanning_configuration` ❌
- `aws_ecr_replication_configuration` ❌
- `aws_ecr_repository_creation_template` ❌
- `aws_ecr_repository_policy` ❌

#### ECRPUBLIC
- `aws_ecrpublic_repository_policy` ❌

#### ECS
- `aws_ecs_account_setting_default` ❌
- `aws_ecs_cluster_capacity_providers` ❌
- `aws_ecs_tag` ❌

#### EFS
- `aws_efs_backup_policy` ❌
- `aws_efs_file_system_policy` ❌
- `aws_efs_mount_target` ❌
- `aws_efs_replication_configuration` ❌

#### EIP
- `aws_eip_association` ❌
- `aws_eip_domain_name` ❌

#### EKS
- `aws_eks_access_policy_association` ❌

#### ELASTIC
- `aws_elastic_beanstalk_configuration_template` ❌

#### ELASTICACHE
- `aws_elasticache_global_replication_group` ❌
- `aws_elasticache_user_group_association` ❌

#### ELASTICSEARCH
- `aws_elasticsearch_domain_policy` ❌
- `aws_elasticsearch_domain_saml_options` ❌
- `aws_elasticsearch_vpc_endpoint` ❌

#### ELASTICTRANSCODER
- `aws_elastictranscoder_pipeline` ❌
- `aws_elastictranscoder_preset` ❌

#### ELB
- `aws_elb_attachment` ❌

#### EMR
- `aws_emr_block_public_access_configuration` ❌
- `aws_emr_instance_fleet` ❌
- `aws_emr_instance_group` ❌
- `aws_emr_managed_scaling_policy` ❌
- `aws_emr_security_configuration` ❌
- `aws_emr_studio_session_mapping` ❌

#### FMS
- `aws_fms_admin_account` ❌

#### GLACIER
- `aws_glacier_vault_lock` ❌

#### GLOBALACCELERATOR
- `aws_globalaccelerator_custom_routing_endpoint_group` ❌
- `aws_globalaccelerator_custom_routing_listener` ❌
- `aws_globalaccelerator_endpoint_group` ❌
- `aws_globalaccelerator_listener` ❌

#### GLUE
- `aws_glue_catalog_table` ❌
- `aws_glue_catalog_table_optimizer` ❌
- `aws_glue_classifier` ❌
- `aws_glue_data_catalog_encryption_settings` ❌
- `aws_glue_partition` ❌
- `aws_glue_partition_index` ❌
- `aws_glue_resource_policy` ❌
- `aws_glue_security_configuration` ❌
- `aws_glue_user_defined_function` ❌

#### GRAFANA
- `aws_grafana_license_association` ❌
- `aws_grafana_role_association` ❌
- `aws_grafana_workspace_api_key` ❌
- `aws_grafana_workspace_saml_configuration` ❌
- `aws_grafana_workspace_service_account` ❌
- `aws_grafana_workspace_service_account_token` ❌

#### GUARDDUTY
- `aws_guardduty_detector_feature` ❌
- `aws_guardduty_invite_accepter` ❌
- `aws_guardduty_member` ❌
- `aws_guardduty_member_detector_feature` ❌
- `aws_guardduty_organization_admin_account` ❌
- `aws_guardduty_organization_configuration` ❌
- `aws_guardduty_organization_configuration_feature` ❌
- `aws_guardduty_publishing_destination` ❌

#### IAM
- `aws_iam_access_key` ❌
- `aws_iam_account_alias` ❌
- `aws_iam_account_password_policy` ❌
- `aws_iam_group` ❌
- `aws_iam_group_membership` ❌
- `aws_iam_group_policies_exclusive` ❌
- `aws_iam_group_policy` ❌
- `aws_iam_group_policy_attachment` ❌
- `aws_iam_group_policy_attachments_exclusive` ❌
- `aws_iam_organizations_features` ❌
- `aws_iam_policy_attachment` ❌
- `aws_iam_role_policies_exclusive` ❌
- `aws_iam_role_policy` ❌
- `aws_iam_role_policy_attachment` ❌
- `aws_iam_role_policy_attachments_exclusive` ❌
- `aws_iam_security_token_service_preferences` ❌
- `aws_iam_service_specific_credential` ❌
- `aws_iam_signing_certificate` ❌
- `aws_iam_user_group_membership` ❌
- `aws_iam_user_login_profile` ❌
- `aws_iam_user_policies_exclusive` ❌
- `aws_iam_user_policy` ❌
- `aws_iam_user_policy_attachment` ❌
- `aws_iam_user_policy_attachments_exclusive` ❌
- `aws_iam_user_ssh_key` ❌

#### IDENTITYSTORE
- `aws_identitystore_group` ❌
- `aws_identitystore_group_membership` ❌
- `aws_identitystore_user` ❌

#### INSPECTOR
- `aws_inspector_assessment_target` ❌

#### INSPECTOR2
- `aws_inspector2_delegated_admin_account` ❌
- `aws_inspector2_enabler` ❌
- `aws_inspector2_member_association` ❌
- `aws_inspector2_organization_configuration` ❌

#### INTERNET
- `aws_internet_gateway_attachment` ❌

#### IOT
- `aws_iot_certificate` ❌
- `aws_iot_event_configurations` ❌
- `aws_iot_indexing_configuration` ❌
- `aws_iot_logging_options` ❌
- `aws_iot_policy_attachment` ❌
- `aws_iot_thing` ❌
- `aws_iot_thing_group_membership` ❌
- `aws_iot_thing_principal_attachment` ❌
- `aws_iot_topic_rule_destination` ❌

#### KENDRA
- `aws_kendra_experience` ❌

#### KINESIS
- `aws_kinesis_resource_policy` ❌
- `aws_kinesis_stream_consumer` ❌

#### KINESISANALYTICSV2
- `aws_kinesisanalyticsv2_application_snapshot` ❌

#### KMS
- `aws_kms_alias` ❌
- `aws_kms_ciphertext` ❌
- `aws_kms_custom_key_store` ❌
- `aws_kms_grant` ❌
- `aws_kms_key_policy` ❌

#### LAKEFORMATION
- `aws_lakeformation_data_cells_filter` ❌
- `aws_lakeformation_data_lake_settings` ❌
- `aws_lakeformation_lf_tag` ❌
- `aws_lakeformation_opt_in` ❌
- `aws_lakeformation_permissions` ❌
- `aws_lakeformation_resource` ❌
- `aws_lakeformation_resource_lf_tag` ❌
- `aws_lakeformation_resource_lf_tags` ❌

#### LAMBDA
- `aws_lambda_alias` ❌
- `aws_lambda_function_event_invoke_config` ❌
- `aws_lambda_function_recursion_config` ❌
- `aws_lambda_function_url` ❌
- `aws_lambda_invocation` ❌
- `aws_lambda_layer_version` ❌
- `aws_lambda_layer_version_permission` ❌
- `aws_lambda_permission` ❌
- `aws_lambda_provisioned_concurrency_config` ❌
- `aws_lambda_runtime_management_config` ❌

#### LAUNCH
- `aws_launch_configuration` ❌

#### LB
- `aws_lb_cookie_stickiness_policy` ❌
- `aws_lb_listener_certificate` ❌
- `aws_lb_ssl_negotiation_policy` ❌
- `aws_lb_target_group_attachment` ❌
- `aws_lb_trust_store_revocation` ❌

#### LEX
- `aws_lex_bot` ❌
- `aws_lex_bot_alias` ❌
- `aws_lex_intent` ❌
- `aws_lex_slot_type` ❌

#### LEXV2MODELS
- `aws_lexv2models_bot_locale` ❌
- `aws_lexv2models_bot_version` ❌
- `aws_lexv2models_intent` ❌
- `aws_lexv2models_slot` ❌
- `aws_lexv2models_slot_type` ❌

#### LICENSEMANAGER
- `aws_licensemanager_association` ❌
- `aws_licensemanager_grant` ❌
- `aws_licensemanager_grant_accepter` ❌

#### LIGHTSAIL
- `aws_lightsail_bucket_access_key` ❌
- `aws_lightsail_bucket_resource_access` ❌
- `aws_lightsail_container_service_deployment_version` ❌
- `aws_lightsail_disk_attachment` ❌
- `aws_lightsail_domain` ❌
- `aws_lightsail_domain_entry` ❌
- `aws_lightsail_instance_public_ports` ❌
- `aws_lightsail_lb_attachment` ❌
- `aws_lightsail_lb_certificate` ❌
- `aws_lightsail_lb_certificate_attachment` ❌
- `aws_lightsail_lb_https_redirection_policy` ❌
- `aws_lightsail_lb_stickiness_policy` ❌
- `aws_lightsail_static_ip` ❌
- `aws_lightsail_static_ip_attachment` ❌

#### LOAD
- `aws_load_balancer_backend_server_policy` ❌
- `aws_load_balancer_listener_policy` ❌
- `aws_load_balancer_policy` ❌

#### LOCATION
- `aws_location_tracker_association` ❌

#### M2
- `aws_m2_deployment` ❌

#### MACIE2
- `aws_macie2_account` ❌
- `aws_macie2_classification_export_configuration` ❌
- `aws_macie2_invitation_accepter` ❌
- `aws_macie2_organization_admin_account` ❌
- `aws_macie2_organization_configuration` ❌

#### MAIN
- `aws_main_route_table_association` ❌

#### MEDIA
- `aws_media_store_container_policy` ❌

#### MEDIALIVE
- `aws_medialive_multiplex_program` ❌

#### MSK
- `aws_msk_cluster_policy` ❌
- `aws_msk_configuration` ❌
- `aws_msk_scram_secret_association` ❌
- `aws_msk_single_scram_secret_association` ❌

#### NEPTUNE
- `aws_neptune_cluster_snapshot` ❌
- `aws_neptune_global_cluster` ❌

#### NETWORK
- `aws_network_acl_association` ❌
- `aws_network_acl_rule` ❌
- `aws_network_interface_attachment` ❌
- `aws_network_interface_permission` ❌
- `aws_network_interface_sg_attachment` ❌

#### NETWORKFIREWALL
- `aws_networkfirewall_logging_configuration` ❌
- `aws_networkfirewall_resource_policy` ❌

#### NETWORKMANAGER
- `aws_networkmanager_attachment_accepter` ❌
- `aws_networkmanager_core_network_policy_attachment` ❌
- `aws_networkmanager_customer_gateway_association` ❌
- `aws_networkmanager_link_association` ❌
- `aws_networkmanager_transit_gateway_connect_peer_association` ❌
- `aws_networkmanager_transit_gateway_registration` ❌

#### NOTIFICATIONS
- `aws_notifications_channel_association` ❌
- `aws_notifications_event_rule` ❌
- `aws_notifications_notification_hub` ❌

#### OAM
- `aws_oam_sink_policy` ❌

#### OPENSEARCH
- `aws_opensearch_authorize_vpc_endpoint_access` ❌
- `aws_opensearch_domain_policy` ❌
- `aws_opensearch_domain_saml_options` ❌
- `aws_opensearch_inbound_connection_accepter` ❌
- `aws_opensearch_outbound_connection` ❌
- `aws_opensearch_package` ❌
- `aws_opensearch_package_association` ❌
- `aws_opensearch_vpc_endpoint` ❌

#### OPENSEARCHSERVERLESS
- `aws_opensearchserverless_access_policy` ❌
- `aws_opensearchserverless_lifecycle_policy` ❌
- `aws_opensearchserverless_security_config` ❌
- `aws_opensearchserverless_security_policy` ❌
- `aws_opensearchserverless_vpc_endpoint` ❌

#### ORGANIZATIONS
- `aws_organizations_delegated_administrator` ❌
- `aws_organizations_organization` ❌
- `aws_organizations_policy_attachment` ❌

#### PAYMENTCRYPTOGRAPHY
- `aws_paymentcryptography_key_alias` ❌

#### PINPOINT
- `aws_pinpoint_adm_channel` ❌
- `aws_pinpoint_apns_channel` ❌
- `aws_pinpoint_apns_sandbox_channel` ❌
- `aws_pinpoint_apns_voip_channel` ❌
- `aws_pinpoint_apns_voip_sandbox_channel` ❌
- `aws_pinpoint_baidu_channel` ❌
- `aws_pinpoint_email_channel` ❌
- `aws_pinpoint_event_stream` ❌
- `aws_pinpoint_gcm_channel` ❌
- `aws_pinpoint_sms_channel` ❌

#### PROMETHEUS
- `aws_prometheus_alert_manager_definition` ❌
- `aws_prometheus_workspace_configuration` ❌

#### PROXY
- `aws_proxy_protocol_policy` ❌

#### QUICKSIGHT
- `aws_quicksight_account_settings` ❌
- `aws_quicksight_account_subscription` ❌
- `aws_quicksight_folder_membership` ❌
- `aws_quicksight_group` ❌
- `aws_quicksight_group_membership` ❌
- `aws_quicksight_iam_policy_assignment` ❌
- `aws_quicksight_ingestion` ❌
- `aws_quicksight_refresh_schedule` ❌
- `aws_quicksight_role_membership` ❌
- `aws_quicksight_template_alias` ❌
- `aws_quicksight_user` ❌

#### RAM
- `aws_ram_principal_association` ❌
- `aws_ram_resource_association` ❌
- `aws_ram_resource_share_accepter` ❌
- `aws_ram_sharing_with_organization` ❌

#### RDS
- `aws_rds_certificate` ❌
- `aws_rds_cluster_activity_stream` ❌
- `aws_rds_cluster_role_association` ❌
- `aws_rds_export_task` ❌
- `aws_rds_instance_state` ❌

#### REDSHIFT
- `aws_redshift_authentication_profile` ❌
- `aws_redshift_cluster_iam_roles` ❌
- `aws_redshift_data_share_authorization` ❌
- `aws_redshift_data_share_consumer_association` ❌
- `aws_redshift_endpoint_access` ❌
- `aws_redshift_endpoint_authorization` ❌
- `aws_redshift_logging` ❌
- `aws_redshift_partner` ❌
- `aws_redshift_resource_policy` ❌
- `aws_redshift_scheduled_action` ❌
- `aws_redshift_snapshot_copy` ❌
- `aws_redshift_snapshot_schedule_association` ❌

#### REDSHIFTDATA
- `aws_redshiftdata_statement` ❌

#### REDSHIFTSERVERLESS
- `aws_redshiftserverless_custom_domain_association` ❌
- `aws_redshiftserverless_endpoint_access` ❌
- `aws_redshiftserverless_resource_policy` ❌
- `aws_redshiftserverless_snapshot` ❌
- `aws_redshiftserverless_usage_limit` ❌

#### RESOURCEGROUPS
- `aws_resourcegroups_resource` ❌

#### ROUTE
- `aws_route` ❌
- `aws_route_table_association` ❌

#### ROUTE53
- `aws_route53_cidr_collection` ❌
- `aws_route53_cidr_location` ❌
- `aws_route53_delegation_set` ❌
- `aws_route53_hosted_zone_dnssec` ❌
- `aws_route53_key_signing_key` ❌
- `aws_route53_query_log` ❌
- `aws_route53_record` ❌
- `aws_route53_records_exclusive` ❌
- `aws_route53_resolver_config` ❌
- `aws_route53_resolver_dnssec_config` ❌
- `aws_route53_resolver_firewall_config` ❌
- `aws_route53_resolver_firewall_rule` ❌
- `aws_route53_resolver_query_log_config_association` ❌
- `aws_route53_resolver_rule_association` ❌
- `aws_route53_traffic_policy` ❌
- `aws_route53_traffic_policy_instance` ❌
- `aws_route53_vpc_association_authorization` ❌
- `aws_route53_zone_association` ❌

#### ROUTE53DOMAINS
- `aws_route53domains_delegation_signer_record` ❌

#### ROUTE53PROFILES
- `aws_route53profiles_resource_association` ❌

#### ROUTE53RECOVERYCONTROLCONFIG
- `aws_route53recoverycontrolconfig_cluster` ❌
- `aws_route53recoverycontrolconfig_control_panel` ❌
- `aws_route53recoverycontrolconfig_routing_control` ❌
- `aws_route53recoverycontrolconfig_safety_rule` ❌

#### RUM
- `aws_rum_metrics_destination` ❌

#### S3
- `aws_s3_access_point` ❌
- `aws_s3_account_public_access_block` ❌
- `aws_s3_bucket_accelerate_configuration` ❌
- `aws_s3_bucket_acl` ❌
- `aws_s3_bucket_analytics_configuration` ❌
- `aws_s3_bucket_cors_configuration` ❌
- `aws_s3_bucket_intelligent_tiering_configuration` ❌
- `aws_s3_bucket_inventory` ❌
- `aws_s3_bucket_lifecycle_configuration` ❌
- `aws_s3_bucket_logging` ❌
- `aws_s3_bucket_metric` ❌
- `aws_s3_bucket_notification` ❌
- `aws_s3_bucket_object_lock_configuration` ❌
- `aws_s3_bucket_ownership_controls` ❌
- `aws_s3_bucket_policy` ❌
- `aws_s3_bucket_public_access_block` ❌
- `aws_s3_bucket_replication_configuration` ❌
- `aws_s3_bucket_request_payment_configuration` ❌
- `aws_s3_bucket_server_side_encryption_configuration` ❌
- `aws_s3_bucket_versioning` ❌
- `aws_s3_bucket_website_configuration` ❌
- `aws_s3_directory_bucket` ❌

#### S3CONTROL
- `aws_s3control_access_grants_instance_resource_policy` ❌
- `aws_s3control_access_point_policy` ❌
- `aws_s3control_bucket_lifecycle_configuration` ❌
- `aws_s3control_bucket_policy` ❌
- `aws_s3control_directory_bucket_access_point_scope` ❌
- `aws_s3control_multi_region_access_point` ❌
- `aws_s3control_multi_region_access_point_policy` ❌
- `aws_s3control_object_lambda_access_point` ❌
- `aws_s3control_object_lambda_access_point_policy` ❌

#### S3OUTPOSTS
- `aws_s3outposts_endpoint` ❌

#### S3TABLES
- `aws_s3tables_namespace` ❌
- `aws_s3tables_table` ❌
- `aws_s3tables_table_bucket` ❌
- `aws_s3tables_table_bucket_policy` ❌
- `aws_s3tables_table_policy` ❌

#### SAGEMAKER
- `aws_sagemaker_device` ❌
- `aws_sagemaker_image_version` ❌
- `aws_sagemaker_model_package_group_policy` ❌
- `aws_sagemaker_servicecatalog_portfolio_status` ❌
- `aws_sagemaker_workforce` ❌

#### SCHEDULER
- `aws_scheduler_schedule` ❌

#### SCHEMAS
- `aws_schemas_registry_policy` ❌

#### SECRETSMANAGER
- `aws_secretsmanager_secret_policy` ❌
- `aws_secretsmanager_secret_rotation` ❌
- `aws_secretsmanager_secret_version` ❌

#### SECURITY
- `aws_security_group_rule` ❌

#### SECURITYHUB
- `aws_securityhub_account` ❌
- `aws_securityhub_action_target` ❌
- `aws_securityhub_configuration_policy` ❌
- `aws_securityhub_configuration_policy_association` ❌
- `aws_securityhub_finding_aggregator` ❌
- `aws_securityhub_insight` ❌
- `aws_securityhub_invite_accepter` ❌
- `aws_securityhub_member` ❌
- `aws_securityhub_organization_admin_account` ❌
- `aws_securityhub_organization_configuration` ❌
- `aws_securityhub_product_subscription` ❌
- `aws_securityhub_standards_control` ❌
- `aws_securityhub_standards_control_association` ❌
- `aws_securityhub_standards_subscription` ❌

#### SECURITYLAKE
- `aws_securitylake_aws_log_source` ❌
- `aws_securitylake_custom_log_source` ❌
- `aws_securitylake_subscriber_notification` ❌

#### SERVICE
- `aws_service_discovery_instance` ❌

#### SERVICECATALOG
- `aws_servicecatalog_budget_resource_association` ❌
- `aws_servicecatalog_constraint` ❌
- `aws_servicecatalog_organizations_access` ❌
- `aws_servicecatalog_portfolio_share` ❌
- `aws_servicecatalog_principal_portfolio_association` ❌
- `aws_servicecatalog_product_portfolio_association` ❌
- `aws_servicecatalog_provisioning_artifact` ❌
- `aws_servicecatalog_service_action` ❌
- `aws_servicecatalog_tag_option` ❌
- `aws_servicecatalog_tag_option_resource_association` ❌

#### SERVICECATALOGAPPREGISTRY
- `aws_servicecatalogappregistry_attribute_group_association` ❌

#### SERVICEQUOTAS
- `aws_servicequotas_service_quota` ❌
- `aws_servicequotas_template` ❌
- `aws_servicequotas_template_association` ❌

#### SES
- `aws_ses_active_receipt_rule_set` ❌
- `aws_ses_configuration_set` ❌
- `aws_ses_domain_dkim` ❌
- `aws_ses_domain_identity` ❌
- `aws_ses_domain_identity_verification` ❌
- `aws_ses_domain_mail_from` ❌
- `aws_ses_email_identity` ❌
- `aws_ses_event_destination` ❌
- `aws_ses_identity_notification_topic` ❌
- `aws_ses_identity_policy` ❌
- `aws_ses_receipt_filter` ❌
- `aws_ses_receipt_rule` ❌
- `aws_ses_receipt_rule_set` ❌
- `aws_ses_template` ❌

#### SESV2
- `aws_sesv2_account_suppression_attributes` ❌
- `aws_sesv2_account_vdm_attributes` ❌
- `aws_sesv2_configuration_set_event_destination` ❌
- `aws_sesv2_dedicated_ip_assignment` ❌
- `aws_sesv2_email_identity_feedback_attributes` ❌
- `aws_sesv2_email_identity_mail_from_attributes` ❌
- `aws_sesv2_email_identity_policy` ❌

#### SFN
- `aws_sfn_alias` ❌

#### SHIELD
- `aws_shield_application_layer_automatic_response` ❌
- `aws_shield_drt_access_log_bucket_association` ❌
- `aws_shield_drt_access_role_arn_association` ❌
- `aws_shield_proactive_engagement` ❌
- `aws_shield_protection_health_check_association` ❌
- `aws_shield_subscription` ❌

#### SIGNER
- `aws_signer_signing_job` ❌
- `aws_signer_signing_profile_permission` ❌

#### SNAPSHOT
- `aws_snapshot_create_volume_permission` ❌

#### SNS
- `aws_sns_platform_application` ❌
- `aws_sns_sms_preferences` ❌
- `aws_sns_topic_data_protection_policy` ❌
- `aws_sns_topic_policy` ❌
- `aws_sns_topic_subscription` ❌

#### SPOT
- `aws_spot_datafeed_subscription` ❌

#### SQS
- `aws_sqs_queue_policy` ❌
- `aws_sqs_queue_redrive_allow_policy` ❌
- `aws_sqs_queue_redrive_policy` ❌

#### SSM
- `aws_ssm_default_patch_baseline` ❌
- `aws_ssm_maintenance_window_target` ❌
- `aws_ssm_maintenance_window_task` ❌
- `aws_ssm_patch_group` ❌
- `aws_ssm_resource_data_sync` ❌
- `aws_ssm_service_setting` ❌

#### SSMCONTACTS
- `aws_ssmcontacts_contact_channel` ❌
- `aws_ssmcontacts_plan` ❌

#### SSOADMIN
- `aws_ssoadmin_account_assignment` ❌
- `aws_ssoadmin_application_access_scope` ❌
- `aws_ssoadmin_application_assignment` ❌
- `aws_ssoadmin_application_assignment_configuration` ❌
- `aws_ssoadmin_customer_managed_policy_attachment` ❌
- `aws_ssoadmin_instance_access_control_attributes` ❌
- `aws_ssoadmin_managed_policy_attachment` ❌
- `aws_ssoadmin_permission_set_inline_policy` ❌
- `aws_ssoadmin_permissions_boundary_attachment` ❌

#### STORAGEGATEWAY
- `aws_storagegateway_cache` ❌
- `aws_storagegateway_upload_buffer` ❌
- `aws_storagegateway_working_storage` ❌

#### SYNTHETICS
- `aws_synthetics_group_association` ❌

#### TRANSFER
- `aws_transfer_access` ❌
- `aws_transfer_ssh_key` ❌
- `aws_transfer_tag` ❌

#### VERIFIEDACCESS
- `aws_verifiedaccess_instance_logging_configuration` ❌
- `aws_verifiedaccess_instance_trust_provider_attachment` ❌

#### VERIFIEDPERMISSIONS
- `aws_verifiedpermissions_identity_source` ❌
- `aws_verifiedpermissions_policy` ❌
- `aws_verifiedpermissions_policy_template` ❌
- `aws_verifiedpermissions_schema` ❌

#### VOLUME
- `aws_volume_attachment` ❌

#### VPC
- `aws_vpc_block_public_access_options` ❌
- `aws_vpc_dhcp_options_association` ❌
- `aws_vpc_endpoint_connection_accepter` ❌
- `aws_vpc_endpoint_connection_notification` ❌
- `aws_vpc_endpoint_policy` ❌
- `aws_vpc_endpoint_private_dns` ❌
- `aws_vpc_endpoint_route_table_association` ❌
- `aws_vpc_endpoint_security_group_association` ❌
- `aws_vpc_endpoint_service_allowed_principal` ❌
- `aws_vpc_endpoint_service_private_dns_verification` ❌
- `aws_vpc_endpoint_subnet_association` ❌
- `aws_vpc_ipam_organization_admin_account` ❌
- `aws_vpc_ipam_pool_cidr` ❌
- `aws_vpc_ipam_pool_cidr_allocation` ❌
- `aws_vpc_ipam_preview_next_cidr` ❌
- `aws_vpc_ipv4_cidr_block_association` ❌
- `aws_vpc_ipv6_cidr_block_association` ❌
- `aws_vpc_network_performance_metric_subscription` ❌
- `aws_vpc_peering_connection_options` ❌
- `aws_vpc_route_server_propagation` ❌
- `aws_vpc_route_server_vpc_association` ❌
- `aws_vpc_security_group_vpc_association` ❌

#### VPCLATTICE
- `aws_vpclattice_auth_policy` ❌
- `aws_vpclattice_resource_policy` ❌
- `aws_vpclattice_target_group_attachment` ❌

#### VPN
- `aws_vpn_connection_route` ❌
- `aws_vpn_gateway_attachment` ❌
- `aws_vpn_gateway_route_propagation` ❌

#### WAF
- `aws_waf_byte_match_set` ❌
- `aws_waf_geo_match_set` ❌
- `aws_waf_ipset` ❌
- `aws_waf_regex_match_set` ❌
- `aws_waf_regex_pattern_set` ❌
- `aws_waf_size_constraint_set` ❌
- `aws_waf_sql_injection_match_set` ❌
- `aws_waf_xss_match_set` ❌

#### WAFREGIONAL
- `aws_wafregional_byte_match_set` ❌
- `aws_wafregional_geo_match_set` ❌
- `aws_wafregional_ipset` ❌
- `aws_wafregional_regex_match_set` ❌
- `aws_wafregional_regex_pattern_set` ❌
- `aws_wafregional_size_constraint_set` ❌
- `aws_wafregional_sql_injection_match_set` ❌
- `aws_wafregional_web_acl_association` ❌
- `aws_wafregional_xss_match_set` ❌

#### WAFV2
- `aws_wafv2_api_key` ❌
- `aws_wafv2_web_acl_association` ❌
- `aws_wafv2_web_acl_logging_configuration` ❌

#### XRAY
- `aws_xray_encryption_config` ❌
- `aws_xray_resource_policy` ❌

## Tagging Patterns and Insights

### Key Observations

1. **Configuration Resources**: Most configuration-only resources (like policies, rules, attachments) don't support tagging
2. **Core Infrastructure**: Primary infrastructure resources (instances, volumes, networks) typically support tagging
3. **Relationship Resources**: Resources that establish relationships between other resources often don't support tagging
4. **Permission Resources**: IAM policies, roles, and permission resources have mixed tagging support

### Services with High Tagging Coverage

The following services have excellent tagging support (>80% of resources):

- **appmesh**: 100.0% (7/7 resources)
- **datasync**: 100.0% (13/13 resources)
- **default**: 100.0% (6/6 resources)
- **dms**: 100.0% (8/8 resources)
- **finspace**: 100.0% (7/7 resources)
- **fsx**: 100.0% (11/11 resources)
- **gamelift**: 100.0% (6/6 resources)
- **imagebuilder**: 100.0% (9/9 resources)
- **memorydb**: 100.0% (7/7 resources)
- **workspacesweb**: 100.0% (6/6 resources)
- **eks**: 87.5% (7/8 resources)
- **devicefarm**: 83.3% (5/6 resources)
- **kendra**: 83.3% (5/6 resources)
- **location**: 83.3% (5/6 resources)
- **sagemaker**: 83.3% (25/30 resources)
- **appfabric**: 80.0% (4/5 resources)
- **elasticache**: 80.0% (8/10 resources)
- **media**: 80.0% (4/5 resources)
- **medialive**: 80.0% (4/5 resources)
- **service**: 80.0% (4/5 resources)

### Services with Limited Tagging Support

Services with low tagging coverage (<30% of resources):

- **autoscaling**: 0.0% (0/8 resources)
- **lakeformation**: 0.0% (0/8 resources)
- **s3tables**: 0.0% (0/5 resources)
- **ses**: 0.0% (0/14 resources)
- **securityhub**: 6.7% (1/15 resources)
- **appsync**: 10.0% (1/10 resources)
- **datazone**: 10.0% (1/10 resources)
- **ecr**: 11.1% (1/9 resources)
- **opensearch**: 11.1% (1/9 resources)
- **cloudfront**: 12.5% (2/16 resources)
- **chime**: 14.3% (1/7 resources)
- **cognito**: 14.3% (2/14 resources)
- **grafana**: 14.3% (1/7 resources)
- **s3**: 15.4% (4/26 resources)
- **lexv2models**: 16.7% (1/6 resources)
- **opensearchserverless**: 16.7% (1/6 resources)
- **pinpoint**: 16.7% (2/12 resources)
- **sns**: 16.7% (1/6 resources)
- **acmpca**: 20.0% (1/5 resources)
- **detective**: 20.0% (1/5 resources)
- **inspector2**: 20.0% (1/5 resources)
- **ram**: 20.0% (1/5 resources)
- **verifiedpermissions**: 20.0% (1/5 resources)
- **dynamodb**: 22.2% (2/9 resources)
- **config**: 23.1% (3/13 resources)
- **lambda**: 23.1% (3/13 resources)
- **servicecatalog**: 23.1% (3/13 resources)
- **directory**: 25.0% (2/8 resources)
- **emr**: 25.0% (2/8 resources)
- **shield**: 25.0% (2/8 resources)
- **ssoadmin**: 25.0% (3/12 resources)
- **iam**: 26.5% (9/34 resources)
- **network**: 28.6% (2/7 resources)
- **redshiftserverless**: 28.6% (2/7 resources)

## Planning Your Tagging Strategy

### Recommendations

1. **Focus on Core Resources**: Prioritize tagging for compute, storage, and networking resources
2. **Service-Specific Approach**: Some services have excellent tagging support while others have limited support
3. **Lifecycle Management**: Consider that some resources may not support tagging but are managed through their parent resources
4. **Cost Allocation**: Focus tagging efforts on resources that contribute to AWS costs

### Best Practices

- **Start with High-Impact Resources**: Begin with EC2 instances, EBS volumes, RDS databases, and S3 buckets
- **Use Consistent Naming**: Implement a consistent tagging strategy across all taggable resources
- **Automate Tagging**: Use tools like Terratag to automatically apply tags during infrastructure provisioning
- **Regular Audits**: Regularly review and update your tagging strategy as new resources are added

## Notes

- This data is extracted from the AWS Terraform provider schema
- Tagging support may vary based on AWS region and account settings
- Some resources may support tagging through AWS APIs but not through Terraform
- This information is current as of the generation date and may change with AWS provider updates

---

*Generated from Terratag AWS resource tagging support matrix*
*Last updated: 1,506 resources analyzed*
