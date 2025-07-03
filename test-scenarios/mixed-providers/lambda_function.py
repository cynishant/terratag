#!/usr/bin/env python3
"""
Cross-cloud integration Lambda function for Terratag mixed provider test.
This function demonstrates cross-cloud data synchronization between AWS and GCP.
"""

import json
import boto3
import logging
from google.cloud import storage as gcs
from google.cloud import functions_v1
import os

# Configure logging
logger = logging.getLogger()
logger.setLevel(logging.INFO)

def handler(event, context):
    """
    AWS Lambda handler for cross-cloud data synchronization.
    
    Args:
        event: AWS Lambda event object
        context: AWS Lambda context object
        
    Returns:
        dict: Response object with status and message
    """
    try:
        logger.info(f"Received event: {json.dumps(event)}")
        
        # Get environment variables
        gcp_project_id = os.environ.get('GCP_PROJECT_ID')
        gcp_region = os.environ.get('GCP_REGION')
        s3_bucket = os.environ.get('S3_BUCKET')
        gcs_bucket = os.environ.get('GCS_BUCKET')
        
        # Validate required environment variables
        if not all([gcp_project_id, gcp_region, s3_bucket, gcs_bucket]):
            raise ValueError("Missing required environment variables")
        
        # Initialize AWS clients
        s3_client = boto3.client('s3')
        
        # Initialize GCP clients
        gcs_client = gcs.Client(project=gcp_project_id)
        
        # Example: Sync a configuration file from S3 to GCS
        sync_result = sync_configuration_file(
            s3_client, gcs_client, 
            s3_bucket, gcs_bucket
        )
        
        # Example: Trigger GCP Cloud Function
        trigger_result = trigger_gcp_function(
            gcp_project_id, gcp_region,
            {'source': 'aws-lambda', 'sync_result': sync_result}
        )
        
        response = {
            'statusCode': 200,
            'body': json.dumps({
                'message': 'Cross-cloud sync completed successfully',
                'sync_result': sync_result,
                'trigger_result': trigger_result,
                'environment': {
                    'gcp_project': gcp_project_id,
                    'gcp_region': gcp_region,
                    's3_bucket': s3_bucket,
                    'gcs_bucket': gcs_bucket
                }
            })
        }
        
        logger.info(f"Cross-cloud sync completed: {response}")
        return response
        
    except Exception as e:
        logger.error(f"Error in cross-cloud sync: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps({
                'error': str(e),
                'message': 'Cross-cloud sync failed'
            })
        }

def sync_configuration_file(s3_client, gcs_client, s3_bucket, gcs_bucket):
    """
    Sync a configuration file from S3 to GCS.
    
    Args:
        s3_client: Boto3 S3 client
        gcs_client: Google Cloud Storage client
        s3_bucket: S3 bucket name
        gcs_bucket: GCS bucket name
        
    Returns:
        dict: Sync operation result
    """
    try:
        config_key = 'config/app-config.json'
        
        # Read from S3
        s3_response = s3_client.get_object(Bucket=s3_bucket, Key=config_key)
        config_data = s3_response['Body'].read()
        
        # Write to GCS
        gcs_bucket_obj = gcs_client.bucket(gcs_bucket)
        gcs_blob = gcs_bucket_obj.blob(config_key)
        gcs_blob.upload_from_string(config_data)
        
        logger.info(f"Successfully synced {config_key} from S3 to GCS")
        
        return {
            'status': 'success',
            'file': config_key,
            'size': len(config_data),
            'source': f's3://{s3_bucket}/{config_key}',
            'destination': f'gs://{gcs_bucket}/{config_key}'
        }
        
    except Exception as e:
        logger.error(f"Error syncing configuration file: {str(e)}")
        return {
            'status': 'error',
            'error': str(e)
        }

def trigger_gcp_function(project_id, region, payload):
    """
    Trigger a GCP Cloud Function from AWS Lambda.
    
    Args:
        project_id: GCP project ID
        region: GCP region
        payload: Data to send to the function
        
    Returns:
        dict: Trigger operation result
    """
    try:
        # This is a simplified example - in practice you'd use HTTP trigger
        # or Pub/Sub to communicate with GCP Cloud Functions
        
        logger.info(f"Would trigger GCP function with payload: {payload}")
        
        return {
            'status': 'success',
            'message': f'Triggered GCP function in {project_id}:{region}',
            'payload': payload
        }
        
    except Exception as e:
        logger.error(f"Error triggering GCP function: {str(e)}")
        return {
            'status': 'error',
            'error': str(e)
        }

# For testing purposes
if __name__ == '__main__':
    # Mock event and context for local testing
    test_event = {
        'Records': [{
            'eventSource': 'aws:s3',
            'eventName': 'ObjectCreated:Put',
            's3': {
                'bucket': {'name': 'test-bucket'},
                'object': {'key': 'test-file.txt'}
            }
        }]
    }
    
    class MockContext:
        def __init__(self):
            self.function_name = 'test-function'
            self.function_version = '1'
            self.invoked_function_arn = 'arn:aws:lambda:us-west-2:123456789012:function:test'
            self.memory_limit_in_mb = 256
            self.remaining_time_in_millis = lambda: 30000
    
    # Set test environment variables
    os.environ['GCP_PROJECT_ID'] = 'test-project'
    os.environ['GCP_REGION'] = 'us-central1'
    os.environ['S3_BUCKET'] = 'test-s3-bucket'
    os.environ['GCS_BUCKET'] = 'test-gcs-bucket'
    
    result = handler(test_event, MockContext())
    print(json.dumps(result, indent=2))