resource "archive_file" "on_feed_msg" {
  type        = "zip"
  source_file = "${path.module}/../dist/cmd/on-feed-msg/bootstrap"
  output_path = "${path.module}/../dist/on-feed-msg.zip"
}

resource "aws_lambda_function" "on_feed_msg" {
  function_name    = "on-feed-msg"
  description      = "Email-sub Lambda function for: on-feed-msg"
  role             = aws_iam_role.lambda.arn
  handler          = "bootstrap"
  filename         = archive_file.on_feed_msg.output_path
  source_code_hash = archive_file.on_feed_msg.output_base64sha256
  runtime          = "provided.al2023"
  architectures    = ["arm64"]
  timeout          = 10
  logging_config {
    log_format = "Text"
    log_group  = aws_cloudwatch_log_group.on_feed_msg.name
  }
  environment {
    variables = {
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.email_sub.name
      SQS_QUEUE_URL       = aws_sqs_queue.email_sub_posts.url
    }
  }

  depends_on = [
    aws_iam_role.lambda,
    aws_iam_role_policy_attachment.lambda_dynamodb,
    aws_iam_role_policy_attachment.lambda_logging,
    aws_dynamodb_table.email_sub,
    aws_sqs_queue.email_sub_feeds,
    aws_sqs_queue.email_sub_posts,
    aws_cloudwatch_log_group.on_feed_msg,
  ]
}

resource "aws_lambda_event_source_mapping" "on_feed_msg" {
  event_source_arn = aws_sqs_queue.email_sub_feeds.arn
  function_name    = aws_lambda_function.on_feed_msg.arn
  batch_size       = 1
  enabled          = true

  depends_on = [
    aws_lambda_function.on_feed_msg,
    aws_sqs_queue.email_sub_feeds,
  ]
}
