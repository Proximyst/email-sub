resource "archive_file" "on_timer" {
  type        = "zip"
  source_file = "${path.module}/../dist/cmd/on-timer/bootstrap"
  output_path = "${path.module}/../dist/on-timer.zip"
}

resource "aws_lambda_function" "on_timer" {
  function_name    = "on-timer"
  description      = "Email-sub Lambda function for: on-timer"
  role             = aws_iam_role.lambda.arn
  handler          = "bootstrap"
  filename         = archive_file.on_timer.output_path
  source_code_hash = archive_file.on_timer.output_base64sha256
  runtime          = "provided.al2023"
  architectures    = ["arm64"]
  logging_config {
    log_format = "Text"
    log_group  = aws_cloudwatch_log_group.on_timer.name
  }
  environment {
    variables = {
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.email_sub.name
      SQS_QUEUE_URL       = aws_sqs_queue.email_sub_feeds.url
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.on_timer,
    aws_iam_role.lambda,
    aws_iam_role_policy_attachment.lambda_dynamodb,
    aws_iam_role_policy_attachment.lambda_logging,
    aws_dynamodb_table.email_sub,
    aws_sqs_queue.email_sub_feeds,
    aws_cloudwatch_log_group.on_timer,
  ]
}

resource "aws_cloudwatch_event_rule" "on_timer_trigger" {
  name                = "on-timer-trigger"
  description         = "Trigger for on-timer Lambda function"
  schedule_expression = "rate(15 minutes)"
}

resource "aws_cloudwatch_event_target" "on_timer_trigger" {
  rule      = aws_cloudwatch_event_rule.on_timer_trigger.name
  target_id = "on-timer-trigger"
  arn       = aws_lambda_function.on_timer.arn
}
