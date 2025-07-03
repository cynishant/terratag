package gcp

import (
	"strings"
)

// Generated GCP resource labeling support matrix
// This file is auto-generated from GCP provider schema

var GCPResourceLabelingSupport = map[string]bool{
	"google_ai_platform_notebook_instance": true,
	"google_api_gateway_api": true,
	"google_api_gateway_api_config": true,
	"google_api_gateway_gateway": true,
	"google_apigee_environment": true,
	"google_apigee_instance": true,
	"google_apigee_organization": true,
	"google_app_engine_application": true,
	"google_app_engine_service": true,
	"google_app_engine_version": true,
	"google_artifact_registry_repository": true,
	"google_bigquery_dataset": true,
	"google_bigquery_dataset_access": false,
	"google_bigquery_dataset_iam_binding": false,
	"google_bigquery_dataset_iam_member": false,
	"google_bigquery_dataset_iam_policy": false,
	"google_bigquery_table": true,
	"google_bigquery_table_iam_binding": false,
	"google_bigquery_table_iam_member": false,
	"google_bigquery_table_iam_policy": false,
	"google_bigtable_instance": true,
	"google_billing_account_iam_binding": false,
	"google_billing_account_iam_member": false,
	"google_billing_account_iam_policy": false,
	"google_binary_authorization_policy": true,
	"google_cloud_function": true,
	"google_cloud_function_iam_binding": false,
	"google_cloud_function_iam_member": false,
	"google_cloud_function_iam_policy": false,
	"google_cloud_run_domain_mapping": false,
	"google_cloud_run_service": true,
	"google_cloud_run_service_iam_binding": false,
	"google_cloud_run_service_iam_member": false,
	"google_cloud_run_service_iam_policy": false,
	"google_cloud_run_v2_service": true,
	"google_cloud_scheduler_job": true,
	"google_cloud_tasks_queue": true,
	"google_cloudfunctions2_function": true,
	"google_cloudfunctions_function": true,
	"google_cloudfunctions_function_iam_binding": false,
	"google_cloudfunctions_function_iam_member": false,
	"google_cloudfunctions_function_iam_policy": false,
	"google_composer_environment": true,
	"google_compute_address": true,
	"google_compute_attached_disk": false,
	"google_compute_autoscaler": true,
	"google_compute_backend_bucket": false,
	"google_compute_backend_service": true,
	"google_compute_disk": true,
	"google_compute_disk_iam_binding": false,
	"google_compute_disk_iam_member": false,
	"google_compute_disk_iam_policy": false,
	"google_compute_firewall": true,
	"google_compute_firewall_policy": false,
	"google_compute_firewall_policy_association": false,
	"google_compute_firewall_policy_rule": false,
	"google_compute_forwarding_rule": true,
	"google_compute_global_address": true,
	"google_compute_global_forwarding_rule": true,
	"google_compute_health_check": true,
	"google_compute_http_health_check": true,
	"google_compute_https_health_check": true,
	"google_compute_image": true,
	"google_compute_image_iam_binding": false,
	"google_compute_image_iam_member": false,
	"google_compute_image_iam_policy": false,
	"google_compute_instance": true,
	"google_compute_instance_group": true,
	"google_compute_instance_group_manager": true,
	"google_compute_instance_group_named_port": false,
	"google_compute_instance_iam_binding": false,
	"google_compute_instance_iam_member": false,
	"google_compute_instance_iam_policy": false,
	"google_compute_instance_template": true,
	"google_compute_managed_ssl_certificate": true,
	"google_compute_network": true,
	"google_compute_network_peering": false,
	"google_compute_network_peering_routes_config": false,
	"google_compute_organization_security_policy": false,
	"google_compute_organization_security_policy_association": false,
	"google_compute_organization_security_policy_rule": false,
	"google_compute_project_metadata": false,
	"google_compute_project_metadata_item": false,
	"google_compute_region_instance_group_manager": true,
	"google_compute_route": false,
	"google_compute_router": true,
	"google_compute_router_nat": true,
	"google_compute_security_policy": false,
	"google_compute_shared_vpc_host_project": false,
	"google_compute_shared_vpc_service_project": false,
	"google_compute_snapshot": true,
	"google_compute_ssl_certificate": true,
	"google_compute_subnetwork": true,
	"google_compute_subnetwork_iam_binding": false,
	"google_compute_subnetwork_iam_member": false,
	"google_compute_subnetwork_iam_policy": false,
	"google_compute_target_http_proxy": false,
	"google_compute_target_https_proxy": false,
	"google_compute_target_pool": true,
	"google_compute_target_ssl_proxy": false,
	"google_compute_target_tcp_proxy": false,
	"google_compute_url_map": true,
	"google_compute_vpn_gateway": true,
	"google_compute_vpn_tunnel": true,
	"google_container_cluster": true,
	"google_container_cluster_iam_binding": false,
	"google_container_cluster_iam_member": false,
	"google_container_cluster_iam_policy": false,
	"google_container_node_pool": true,
	"google_container_registry": true,
	"google_data_fusion_instance": true,
	"google_dataflow_job": true,
	"google_dataform_repository": true,
	"google_dataplex_asset": true,
	"google_dataplex_lake": true,
	"google_dataplex_zone": true,
	"google_dataproc_cluster": true,
	"google_dns_managed_zone": true,
	"google_dns_policy": false,
	"google_dns_record_set": true,
	"google_endpoints_service": false,
	"google_endpoints_service_iam_binding": false,
	"google_endpoints_service_iam_member": false,
	"google_endpoints_service_iam_policy": false,
	"google_eventarc_trigger": true,
	"google_filestore_instance": true,
	"google_firebase_android_app": true,
	"google_firebase_ios_app": true,
	"google_firebase_project": true,
	"google_firebase_web_app": true,
	"google_folder": false,
	"google_folder_iam_binding": false,
	"google_folder_iam_member": false,
	"google_folder_iam_policy": false,
	"google_healthcare_dataset": true,
	"google_healthcare_dicom_store": true,
	"google_healthcare_fhir_store": true,
	"google_healthcare_hl7_v2_store": true,
	"google_iap_brand": true,
	"google_iap_client": true,
	"google_identity_platform_config": true,
	"google_kms_crypto_key": true,
	"google_kms_crypto_key_iam_binding": false,
	"google_kms_crypto_key_iam_member": false,
	"google_kms_crypto_key_iam_policy": false,
	"google_kms_key_ring": true,
	"google_kms_key_ring_iam_binding": false,
	"google_kms_key_ring_iam_member": false,
	"google_kms_key_ring_iam_policy": false,
	"google_logging_sink": true,
	"google_memcache_instance": true,
	"google_ml_engine_model": true,
	"google_monitoring_alert_policy": true,
	"google_monitoring_notification_channel": true,
	"google_monitoring_uptime_check_config": true,
	"google_network_connectivity_hub": true,
	"google_network_connectivity_spoke": true,
	"google_network_security_gateway_security_policy": true,
	"google_network_security_server_tls_policy": true,
	"google_notebooks_instance": true,
	"google_organization_iam_binding": false,
	"google_organization_iam_member": false,
	"google_organization_iam_policy": false,
	"google_organization_policy": false,
	"google_project": false,
	"google_project_iam_binding": false,
	"google_project_iam_custom_role": false,
	"google_project_iam_member": false,
	"google_project_iam_policy": false,
	"google_project_organization_policy": false,
	"google_project_service": true,
	"google_pubsub_subscription": true,
	"google_pubsub_subscription_iam_binding": false,
	"google_pubsub_subscription_iam_member": false,
	"google_pubsub_subscription_iam_policy": false,
	"google_pubsub_topic": true,
	"google_pubsub_topic_iam_binding": false,
	"google_pubsub_topic_iam_member": false,
	"google_pubsub_topic_iam_policy": false,
	"google_redis_instance": true,
	"google_secret_manager_secret": true,
	"google_secret_manager_secret_iam_binding": false,
	"google_secret_manager_secret_iam_member": false,
	"google_secret_manager_secret_iam_policy": false,
	"google_secret_manager_secret_version": false,
	"google_service_account": true,
	"google_service_account_iam_binding": false,
	"google_service_account_iam_member": false,
	"google_service_account_iam_policy": false,
	"google_service_account_key": false,
	"google_sourcerepo_repository": true,
	"google_spanner_database": true,
	"google_spanner_instance": true,
	"google_sql_database": true,
	"google_sql_database_instance": true,
	"google_sql_ssl_cert": false,
	"google_sql_user": false,
	"google_storage_bucket": true,
	"google_storage_bucket_access_control": false,
	"google_storage_bucket_iam_binding": false,
	"google_storage_bucket_iam_member": false,
	"google_storage_bucket_iam_policy": false,
	"google_storage_bucket_object": true,
	"google_storage_default_object_access_control": false,
	"google_storage_notification": false,
	"google_storage_object_access_control": false,
	"google_vertex_ai_dataset": true,
	"google_vertex_ai_endpoint": true,
	"google_vmwareengine_cluster": true,
	"google_vmwareengine_network": true,
	"google_vmwareengine_private_cloud": true,
	"google_vpc_access_connector": true,
	"google_workflows_workflow": true,
}

// GCPResourceInfo provides detailed information about GCP resource labeling support
type GCPResourceInfo struct {
	ResourceType         string
	SupportsLabels       bool
	LabelAttributeName   string
	Service              string
	Category             string
}

var GCPResourceDetails = map[string]GCPResourceInfo{
	"google_ai_platform_notebook_instance": {
		ResourceType: "google_ai_platform_notebook_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "ai",
		Category: "AI/ML",
	},
	"google_api_gateway_api": {
		ResourceType: "google_api_gateway_api",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "api",
		Category: "API Management",
	},
	"google_api_gateway_api_config": {
		ResourceType: "google_api_gateway_api_config",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "api",
		Category: "API Management",
	},
	"google_api_gateway_gateway": {
		ResourceType: "google_api_gateway_gateway",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "api",
		Category: "API Management",
	},
	"google_apigee_environment": {
		ResourceType: "google_apigee_environment",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "apigee",
		Category: "API Management",
	},
	"google_apigee_instance": {
		ResourceType: "google_apigee_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "apigee",
		Category: "API Management",
	},
	"google_apigee_organization": {
		ResourceType: "google_apigee_organization",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "apigee",
		Category: "API Management",
	},
	"google_app_engine_application": {
		ResourceType: "google_app_engine_application",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "app",
		Category: "App Engine",
	},
	"google_app_engine_service": {
		ResourceType: "google_app_engine_service",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "app",
		Category: "App Engine",
	},
	"google_app_engine_version": {
		ResourceType: "google_app_engine_version",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "app",
		Category: "App Engine",
	},
	"google_artifact_registry_repository": {
		ResourceType: "google_artifact_registry_repository",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "artifact",
		Category: "DevOps",
	},
	"google_bigquery_dataset": {
		ResourceType: "google_bigquery_dataset",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_dataset_access": {
		ResourceType: "google_bigquery_dataset_access",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_dataset_iam_binding": {
		ResourceType: "google_bigquery_dataset_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_dataset_iam_member": {
		ResourceType: "google_bigquery_dataset_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_dataset_iam_policy": {
		ResourceType: "google_bigquery_dataset_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_table": {
		ResourceType: "google_bigquery_table",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_table_iam_binding": {
		ResourceType: "google_bigquery_table_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_table_iam_member": {
		ResourceType: "google_bigquery_table_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigquery_table_iam_policy": {
		ResourceType: "google_bigquery_table_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "bigquery",
		Category: "Analytics",
	},
	"google_bigtable_instance": {
		ResourceType: "google_bigtable_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "bigtable",
		Category: "Database",
	},
	"google_billing_account_iam_binding": {
		ResourceType: "google_billing_account_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "billing",
		Category: "Billing",
	},
	"google_billing_account_iam_member": {
		ResourceType: "google_billing_account_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "billing",
		Category: "Billing",
	},
	"google_billing_account_iam_policy": {
		ResourceType: "google_billing_account_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "billing",
		Category: "Billing",
	},
	"google_binary_authorization_policy": {
		ResourceType: "google_binary_authorization_policy",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "binary",
		Category: "Security",
	},
	"google_cloud_function": {
		ResourceType: "google_cloud_function",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_function_iam_binding": {
		ResourceType: "google_cloud_function_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_function_iam_member": {
		ResourceType: "google_cloud_function_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_function_iam_policy": {
		ResourceType: "google_cloud_function_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_run_domain_mapping": {
		ResourceType: "google_cloud_run_domain_mapping",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_run_service": {
		ResourceType: "google_cloud_run_service",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_run_service_iam_binding": {
		ResourceType: "google_cloud_run_service_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_run_service_iam_member": {
		ResourceType: "google_cloud_run_service_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_run_service_iam_policy": {
		ResourceType: "google_cloud_run_service_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_run_v2_service": {
		ResourceType: "google_cloud_run_v2_service",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_scheduler_job": {
		ResourceType: "google_cloud_scheduler_job",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloud_tasks_queue": {
		ResourceType: "google_cloud_tasks_queue",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "cloud",
		Category: "Serverless",
	},
	"google_cloudfunctions2_function": {
		ResourceType: "google_cloudfunctions2_function",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "cloudfunctions2",
		Category: "Serverless",
	},
	"google_cloudfunctions_function": {
		ResourceType: "google_cloudfunctions_function",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "cloudfunctions",
		Category: "Serverless",
	},
	"google_cloudfunctions_function_iam_binding": {
		ResourceType: "google_cloudfunctions_function_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloudfunctions",
		Category: "Serverless",
	},
	"google_cloudfunctions_function_iam_member": {
		ResourceType: "google_cloudfunctions_function_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloudfunctions",
		Category: "Serverless",
	},
	"google_cloudfunctions_function_iam_policy": {
		ResourceType: "google_cloudfunctions_function_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "cloudfunctions",
		Category: "Serverless",
	},
	"google_composer_environment": {
		ResourceType: "google_composer_environment",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "composer",
		Category: "Analytics",
	},
	"google_compute_address": {
		ResourceType: "google_compute_address",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_attached_disk": {
		ResourceType: "google_compute_attached_disk",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_autoscaler": {
		ResourceType: "google_compute_autoscaler",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_backend_bucket": {
		ResourceType: "google_compute_backend_bucket",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_backend_service": {
		ResourceType: "google_compute_backend_service",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_disk": {
		ResourceType: "google_compute_disk",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_disk_iam_binding": {
		ResourceType: "google_compute_disk_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_disk_iam_member": {
		ResourceType: "google_compute_disk_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_disk_iam_policy": {
		ResourceType: "google_compute_disk_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_firewall": {
		ResourceType: "google_compute_firewall",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_firewall_policy": {
		ResourceType: "google_compute_firewall_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_firewall_policy_association": {
		ResourceType: "google_compute_firewall_policy_association",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_firewall_policy_rule": {
		ResourceType: "google_compute_firewall_policy_rule",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_forwarding_rule": {
		ResourceType: "google_compute_forwarding_rule",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_global_address": {
		ResourceType: "google_compute_global_address",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_global_forwarding_rule": {
		ResourceType: "google_compute_global_forwarding_rule",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_health_check": {
		ResourceType: "google_compute_health_check",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_http_health_check": {
		ResourceType: "google_compute_http_health_check",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_https_health_check": {
		ResourceType: "google_compute_https_health_check",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_image": {
		ResourceType: "google_compute_image",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_image_iam_binding": {
		ResourceType: "google_compute_image_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_image_iam_member": {
		ResourceType: "google_compute_image_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_image_iam_policy": {
		ResourceType: "google_compute_image_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance": {
		ResourceType: "google_compute_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance_group": {
		ResourceType: "google_compute_instance_group",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance_group_manager": {
		ResourceType: "google_compute_instance_group_manager",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance_group_named_port": {
		ResourceType: "google_compute_instance_group_named_port",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance_iam_binding": {
		ResourceType: "google_compute_instance_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance_iam_member": {
		ResourceType: "google_compute_instance_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance_iam_policy": {
		ResourceType: "google_compute_instance_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_instance_template": {
		ResourceType: "google_compute_instance_template",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_managed_ssl_certificate": {
		ResourceType: "google_compute_managed_ssl_certificate",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_network": {
		ResourceType: "google_compute_network",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_network_peering": {
		ResourceType: "google_compute_network_peering",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_network_peering_routes_config": {
		ResourceType: "google_compute_network_peering_routes_config",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_organization_security_policy": {
		ResourceType: "google_compute_organization_security_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_organization_security_policy_association": {
		ResourceType: "google_compute_organization_security_policy_association",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_organization_security_policy_rule": {
		ResourceType: "google_compute_organization_security_policy_rule",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_project_metadata": {
		ResourceType: "google_compute_project_metadata",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_project_metadata_item": {
		ResourceType: "google_compute_project_metadata_item",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_region_instance_group_manager": {
		ResourceType: "google_compute_region_instance_group_manager",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_route": {
		ResourceType: "google_compute_route",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_router": {
		ResourceType: "google_compute_router",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_router_nat": {
		ResourceType: "google_compute_router_nat",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_security_policy": {
		ResourceType: "google_compute_security_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_shared_vpc_host_project": {
		ResourceType: "google_compute_shared_vpc_host_project",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_shared_vpc_service_project": {
		ResourceType: "google_compute_shared_vpc_service_project",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_snapshot": {
		ResourceType: "google_compute_snapshot",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_ssl_certificate": {
		ResourceType: "google_compute_ssl_certificate",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_subnetwork": {
		ResourceType: "google_compute_subnetwork",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_subnetwork_iam_binding": {
		ResourceType: "google_compute_subnetwork_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_subnetwork_iam_member": {
		ResourceType: "google_compute_subnetwork_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_subnetwork_iam_policy": {
		ResourceType: "google_compute_subnetwork_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_target_http_proxy": {
		ResourceType: "google_compute_target_http_proxy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_target_https_proxy": {
		ResourceType: "google_compute_target_https_proxy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_target_pool": {
		ResourceType: "google_compute_target_pool",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_target_ssl_proxy": {
		ResourceType: "google_compute_target_ssl_proxy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_target_tcp_proxy": {
		ResourceType: "google_compute_target_tcp_proxy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_url_map": {
		ResourceType: "google_compute_url_map",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_vpn_gateway": {
		ResourceType: "google_compute_vpn_gateway",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_compute_vpn_tunnel": {
		ResourceType: "google_compute_vpn_tunnel",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "compute",
		Category: "Compute",
	},
	"google_container_cluster": {
		ResourceType: "google_container_cluster",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "container",
		Category: "Kubernetes",
	},
	"google_container_cluster_iam_binding": {
		ResourceType: "google_container_cluster_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "container",
		Category: "Kubernetes",
	},
	"google_container_cluster_iam_member": {
		ResourceType: "google_container_cluster_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "container",
		Category: "Kubernetes",
	},
	"google_container_cluster_iam_policy": {
		ResourceType: "google_container_cluster_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "container",
		Category: "Kubernetes",
	},
	"google_container_node_pool": {
		ResourceType: "google_container_node_pool",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "container",
		Category: "Kubernetes",
	},
	"google_container_registry": {
		ResourceType: "google_container_registry",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "container",
		Category: "Kubernetes",
	},
	"google_data_fusion_instance": {
		ResourceType: "google_data_fusion_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "data",
		Category: "Analytics",
	},
	"google_dataflow_job": {
		ResourceType: "google_dataflow_job",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dataflow",
		Category: "Analytics",
	},
	"google_dataform_repository": {
		ResourceType: "google_dataform_repository",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dataform",
		Category: "Analytics",
	},
	"google_dataplex_asset": {
		ResourceType: "google_dataplex_asset",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dataplex",
		Category: "Analytics",
	},
	"google_dataplex_lake": {
		ResourceType: "google_dataplex_lake",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dataplex",
		Category: "Analytics",
	},
	"google_dataplex_zone": {
		ResourceType: "google_dataplex_zone",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dataplex",
		Category: "Analytics",
	},
	"google_dataproc_cluster": {
		ResourceType: "google_dataproc_cluster",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dataproc",
		Category: "Analytics",
	},
	"google_dns_managed_zone": {
		ResourceType: "google_dns_managed_zone",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dns",
		Category: "Networking",
	},
	"google_dns_policy": {
		ResourceType: "google_dns_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "dns",
		Category: "Networking",
	},
	"google_dns_record_set": {
		ResourceType: "google_dns_record_set",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "dns",
		Category: "Networking",
	},
	"google_endpoints_service": {
		ResourceType: "google_endpoints_service",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "endpoints",
		Category: "API Management",
	},
	"google_endpoints_service_iam_binding": {
		ResourceType: "google_endpoints_service_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "endpoints",
		Category: "API Management",
	},
	"google_endpoints_service_iam_member": {
		ResourceType: "google_endpoints_service_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "endpoints",
		Category: "API Management",
	},
	"google_endpoints_service_iam_policy": {
		ResourceType: "google_endpoints_service_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "endpoints",
		Category: "API Management",
	},
	"google_eventarc_trigger": {
		ResourceType: "google_eventarc_trigger",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "eventarc",
		Category: "Serverless",
	},
	"google_filestore_instance": {
		ResourceType: "google_filestore_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "filestore",
		Category: "Storage",
	},
	"google_firebase_android_app": {
		ResourceType: "google_firebase_android_app",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "firebase",
		Category: "Firebase",
	},
	"google_firebase_ios_app": {
		ResourceType: "google_firebase_ios_app",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "firebase",
		Category: "Firebase",
	},
	"google_firebase_project": {
		ResourceType: "google_firebase_project",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "firebase",
		Category: "Firebase",
	},
	"google_firebase_web_app": {
		ResourceType: "google_firebase_web_app",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "firebase",
		Category: "Firebase",
	},
	"google_folder": {
		ResourceType: "google_folder",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "folder",
		Category: "IAM",
	},
	"google_folder_iam_binding": {
		ResourceType: "google_folder_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "folder",
		Category: "IAM",
	},
	"google_folder_iam_member": {
		ResourceType: "google_folder_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "folder",
		Category: "IAM",
	},
	"google_folder_iam_policy": {
		ResourceType: "google_folder_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "folder",
		Category: "IAM",
	},
	"google_healthcare_dataset": {
		ResourceType: "google_healthcare_dataset",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "healthcare",
		Category: "Healthcare",
	},
	"google_healthcare_dicom_store": {
		ResourceType: "google_healthcare_dicom_store",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "healthcare",
		Category: "Healthcare",
	},
	"google_healthcare_fhir_store": {
		ResourceType: "google_healthcare_fhir_store",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "healthcare",
		Category: "Healthcare",
	},
	"google_healthcare_hl7_v2_store": {
		ResourceType: "google_healthcare_hl7_v2_store",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "healthcare",
		Category: "Healthcare",
	},
	"google_iap_brand": {
		ResourceType: "google_iap_brand",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "iap",
		Category: "Security",
	},
	"google_iap_client": {
		ResourceType: "google_iap_client",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "iap",
		Category: "Security",
	},
	"google_identity_platform_config": {
		ResourceType: "google_identity_platform_config",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "identity",
		Category: "Security",
	},
	"google_kms_crypto_key": {
		ResourceType: "google_kms_crypto_key",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "kms",
		Category: "Security",
	},
	"google_kms_crypto_key_iam_binding": {
		ResourceType: "google_kms_crypto_key_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "kms",
		Category: "Security",
	},
	"google_kms_crypto_key_iam_member": {
		ResourceType: "google_kms_crypto_key_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "kms",
		Category: "Security",
	},
	"google_kms_crypto_key_iam_policy": {
		ResourceType: "google_kms_crypto_key_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "kms",
		Category: "Security",
	},
	"google_kms_key_ring": {
		ResourceType: "google_kms_key_ring",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "kms",
		Category: "Security",
	},
	"google_kms_key_ring_iam_binding": {
		ResourceType: "google_kms_key_ring_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "kms",
		Category: "Security",
	},
	"google_kms_key_ring_iam_member": {
		ResourceType: "google_kms_key_ring_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "kms",
		Category: "Security",
	},
	"google_kms_key_ring_iam_policy": {
		ResourceType: "google_kms_key_ring_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "kms",
		Category: "Security",
	},
	"google_logging_sink": {
		ResourceType: "google_logging_sink",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "logging",
		Category: "Monitoring",
	},
	"google_memcache_instance": {
		ResourceType: "google_memcache_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "memcache",
		Category: "Database",
	},
	"google_ml_engine_model": {
		ResourceType: "google_ml_engine_model",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "ml",
		Category: "AI/ML",
	},
	"google_monitoring_alert_policy": {
		ResourceType: "google_monitoring_alert_policy",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "monitoring",
		Category: "Monitoring",
	},
	"google_monitoring_notification_channel": {
		ResourceType: "google_monitoring_notification_channel",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "monitoring",
		Category: "Monitoring",
	},
	"google_monitoring_uptime_check_config": {
		ResourceType: "google_monitoring_uptime_check_config",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "monitoring",
		Category: "Monitoring",
	},
	"google_network_connectivity_hub": {
		ResourceType: "google_network_connectivity_hub",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "network",
		Category: "Networking",
	},
	"google_network_connectivity_spoke": {
		ResourceType: "google_network_connectivity_spoke",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "network",
		Category: "Networking",
	},
	"google_network_security_gateway_security_policy": {
		ResourceType: "google_network_security_gateway_security_policy",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "network",
		Category: "Networking",
	},
	"google_network_security_server_tls_policy": {
		ResourceType: "google_network_security_server_tls_policy",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "network",
		Category: "Networking",
	},
	"google_notebooks_instance": {
		ResourceType: "google_notebooks_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "notebooks",
		Category: "AI/ML",
	},
	"google_organization_iam_binding": {
		ResourceType: "google_organization_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "organization",
		Category: "IAM",
	},
	"google_organization_iam_member": {
		ResourceType: "google_organization_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "organization",
		Category: "IAM",
	},
	"google_organization_iam_policy": {
		ResourceType: "google_organization_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "organization",
		Category: "IAM",
	},
	"google_organization_policy": {
		ResourceType: "google_organization_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "organization",
		Category: "IAM",
	},
	"google_project": {
		ResourceType: "google_project",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "project",
		Category: "IAM",
	},
	"google_project_iam_binding": {
		ResourceType: "google_project_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "project",
		Category: "IAM",
	},
	"google_project_iam_custom_role": {
		ResourceType: "google_project_iam_custom_role",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "project",
		Category: "IAM",
	},
	"google_project_iam_member": {
		ResourceType: "google_project_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "project",
		Category: "IAM",
	},
	"google_project_iam_policy": {
		ResourceType: "google_project_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "project",
		Category: "IAM",
	},
	"google_project_organization_policy": {
		ResourceType: "google_project_organization_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "project",
		Category: "IAM",
	},
	"google_project_service": {
		ResourceType: "google_project_service",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "project",
		Category: "IAM",
	},
	"google_pubsub_subscription": {
		ResourceType: "google_pubsub_subscription",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_pubsub_subscription_iam_binding": {
		ResourceType: "google_pubsub_subscription_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_pubsub_subscription_iam_member": {
		ResourceType: "google_pubsub_subscription_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_pubsub_subscription_iam_policy": {
		ResourceType: "google_pubsub_subscription_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_pubsub_topic": {
		ResourceType: "google_pubsub_topic",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_pubsub_topic_iam_binding": {
		ResourceType: "google_pubsub_topic_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_pubsub_topic_iam_member": {
		ResourceType: "google_pubsub_topic_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_pubsub_topic_iam_policy": {
		ResourceType: "google_pubsub_topic_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "pubsub",
		Category: "Messaging",
	},
	"google_redis_instance": {
		ResourceType: "google_redis_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "redis",
		Category: "Database",
	},
	"google_secret_manager_secret": {
		ResourceType: "google_secret_manager_secret",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "secret",
		Category: "Security",
	},
	"google_secret_manager_secret_iam_binding": {
		ResourceType: "google_secret_manager_secret_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "secret",
		Category: "Security",
	},
	"google_secret_manager_secret_iam_member": {
		ResourceType: "google_secret_manager_secret_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "secret",
		Category: "Security",
	},
	"google_secret_manager_secret_iam_policy": {
		ResourceType: "google_secret_manager_secret_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "secret",
		Category: "Security",
	},
	"google_secret_manager_secret_version": {
		ResourceType: "google_secret_manager_secret_version",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "secret",
		Category: "Security",
	},
	"google_service_account": {
		ResourceType: "google_service_account",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "service",
		Category: "IAM",
	},
	"google_service_account_iam_binding": {
		ResourceType: "google_service_account_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "service",
		Category: "IAM",
	},
	"google_service_account_iam_member": {
		ResourceType: "google_service_account_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "service",
		Category: "IAM",
	},
	"google_service_account_iam_policy": {
		ResourceType: "google_service_account_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "service",
		Category: "IAM",
	},
	"google_service_account_key": {
		ResourceType: "google_service_account_key",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "service",
		Category: "IAM",
	},
	"google_sourcerepo_repository": {
		ResourceType: "google_sourcerepo_repository",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "sourcerepo",
		Category: "DevOps",
	},
	"google_spanner_database": {
		ResourceType: "google_spanner_database",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "spanner",
		Category: "Database",
	},
	"google_spanner_instance": {
		ResourceType: "google_spanner_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "spanner",
		Category: "Database",
	},
	"google_sql_database": {
		ResourceType: "google_sql_database",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "sql",
		Category: "Database",
	},
	"google_sql_database_instance": {
		ResourceType: "google_sql_database_instance",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "sql",
		Category: "Database",
	},
	"google_sql_ssl_cert": {
		ResourceType: "google_sql_ssl_cert",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "sql",
		Category: "Database",
	},
	"google_sql_user": {
		ResourceType: "google_sql_user",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "sql",
		Category: "Database",
	},
	"google_storage_bucket": {
		ResourceType: "google_storage_bucket",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_bucket_access_control": {
		ResourceType: "google_storage_bucket_access_control",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_bucket_iam_binding": {
		ResourceType: "google_storage_bucket_iam_binding",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_bucket_iam_member": {
		ResourceType: "google_storage_bucket_iam_member",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_bucket_iam_policy": {
		ResourceType: "google_storage_bucket_iam_policy",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_bucket_object": {
		ResourceType: "google_storage_bucket_object",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_default_object_access_control": {
		ResourceType: "google_storage_default_object_access_control",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_notification": {
		ResourceType: "google_storage_notification",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "storage",
		Category: "Storage",
	},
	"google_storage_object_access_control": {
		ResourceType: "google_storage_object_access_control",
		SupportsLabels: false,
		LabelAttributeName: "",
		Service: "storage",
		Category: "Storage",
	},
	"google_vertex_ai_dataset": {
		ResourceType: "google_vertex_ai_dataset",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "vertex",
		Category: "AI/ML",
	},
	"google_vertex_ai_endpoint": {
		ResourceType: "google_vertex_ai_endpoint",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "vertex",
		Category: "AI/ML",
	},
	"google_vmwareengine_cluster": {
		ResourceType: "google_vmwareengine_cluster",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "vmwareengine",
		Category: "VMware Engine",
	},
	"google_vmwareengine_network": {
		ResourceType: "google_vmwareengine_network",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "vmwareengine",
		Category: "VMware Engine",
	},
	"google_vmwareengine_private_cloud": {
		ResourceType: "google_vmwareengine_private_cloud",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "vmwareengine",
		Category: "VMware Engine",
	},
	"google_vpc_access_connector": {
		ResourceType: "google_vpc_access_connector",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "vpc",
		Category: "Networking",
	},
	"google_workflows_workflow": {
		ResourceType: "google_workflows_workflow",
		SupportsLabels: true,
		LabelAttributeName: "labels",
		Service: "workflows",
		Category: "Serverless",
	},
}

// GetGCPLabelingCapability returns labeling information for a GCP resource type
func GetGCPLabelingCapability(resourceType string) (bool, string) {
	if info, exists := GCPResourceDetails[resourceType]; exists {
		return info.SupportsLabels, info.LabelAttributeName
	}
	
	// Default check for google_ prefixed resources
	if strings.HasPrefix(resourceType, "google_") {
		// Most GCP resources support labels, but some configuration resources don't
		if isGCPConfigurationResource(resourceType) {
			return false, ""
		}
		return true, "labels"
	}
	
	return false, ""
}

// isGCPConfigurationResource checks if a resource is a configuration/association resource
func isGCPConfigurationResource(resourceType string) bool {
	configPatterns := []string{
		"_iam_policy",
		"_iam_binding", 
		"_iam_member",
		"_access_control",
		"_peering",
		"_association",
		"_attachment",
		"_policy",
		"_rule",
		"_config",
		"_key",
		"_version",
		"_member",
		"_binding",
		"_metadata",
		"project_service",
		"organization_policy",
		"folder_organization_policy",
	}
	
	for _, pattern := range configPatterns {
		if strings.Contains(resourceType, pattern) {
			return true
		}
	}
	
	return false
}

// GetGCPServiceInfo returns service breakdown information
func GetGCPServiceInfo() map[string]ServiceInfo {
	services := make(map[string][]GCPResourceInfo)
	
	for _, resource := range GCPResourceDetails {
		services[resource.Service] = append(services[resource.Service], resource)
	}
	
	result := make(map[string]ServiceInfo)
	for service, resources := range services {
		total := len(resources)
		labeled := 0
		for _, resource := range resources {
			if resource.SupportsLabels {
				labeled++
			}
		}
		
		result[service] = ServiceInfo{
			TotalResources:   total,
			LabeledResources: labeled,
			LabelingRate:     float64(labeled) / float64(total) * 100,
		}
	}
	
	return result
}

// ServiceInfo contains service-level labeling statistics
type ServiceInfo struct {
	TotalResources   int     `json:"total_resources"`
	LabeledResources int     `json:"labeled_resources"`
	LabelingRate     float64 `json:"labeling_rate"`
}

// SupportsLabeling returns true if the GCP resource supports labels
func SupportsLabeling(resourceType string) bool {
	supported, exists := GCPResourceLabelingSupport[resourceType]
	return exists && supported
}

// GetServiceName extracts the GCP service name from a resource type
func GetServiceName(resourceType string) string {
	if info, exists := GCPResourceDetails[resourceType]; exists {
		return info.Service
	}
	// Fallback extraction
	if strings.HasPrefix(resourceType, "google_") {
		parts := strings.Split(resourceType[7:], "_")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return "unknown"
}

// GetResourceCategory returns the category of a GCP resource
func GetResourceCategory(resourceType string) string {
	if info, exists := GCPResourceDetails[resourceType]; exists {
		return info.Category
	}
	
	if SupportsLabeling(resourceType) {
		return "Compute" // Default category for unknown labeled resources
	}
	
	// Categorize non-labelable resources
	if strings.Contains(resourceType, "_iam_") {
		return "IAM"
	}
	
	if strings.Contains(resourceType, "_access_control") ||
		strings.Contains(resourceType, "_peering") ||
		strings.Contains(resourceType, "_association") ||
		strings.Contains(resourceType, "_attachment") {
		return "Association"
	}
	
	if strings.Contains(resourceType, "_policy") ||
		strings.Contains(resourceType, "_rule") ||
		strings.Contains(resourceType, "_config") {
		return "Configuration"
	}
	
	if strings.Contains(resourceType, "_key") ||
		strings.Contains(resourceType, "_version") ||
		strings.Contains(resourceType, "_member") ||
		strings.Contains(resourceType, "_binding") {
		return "Metadata"
	}
	
	return "Other"
}

// GetLabelingReason provides a human-readable reason for labeling support status
func GetLabelingReason(resourceType string) string {
	if info, exists := GCPResourceDetails[resourceType]; exists {
		if info.SupportsLabels {
			return "Google Cloud resource supports labels"
		}
		
		// Provide specific reasons for non-labelable resources
		if strings.Contains(resourceType, "_iam_") {
			return "IAM resources don't support labels"
		}
		
		if strings.Contains(resourceType, "_access_control") {
			return "Access control resources don't support labels"
		}
		
		if strings.Contains(resourceType, "_policy") {
			return "Policy resources don't support labels"
		}
		
		if strings.Contains(resourceType, "_rule") {
			return "Rule resources don't support labels"
		}
		
		if strings.Contains(resourceType, "_metadata") {
			return "Metadata resources don't support labels"
		}
		
		if strings.Contains(resourceType, "_association") ||
		   strings.Contains(resourceType, "_attachment") ||
		   strings.Contains(resourceType, "_peering") {
			return "Association/attachment resources don't support labels"
		}
		
		return "Configuration or management resource doesn't support labels"
	}
	
	return "Unknown resource type"
}
