data "aws_iam_policy_document" "assume_lambda_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "AssumeLambdaRole"
  description        = "Role for Lambda to assume Lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role.json
}

resource "aws_cloudwatch_log_group" "on_timer" {
  name              = "/aws/lambda/on-timer"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "on_feed_msg" {
  name              = "/aws/lambda/on-feed-msg"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "on_post_msg" {
  name              = "/aws/lambda/on-post-msg"
  retention_in_days = 7
}

data "aws_iam_policy_document" "allow_lambda_logging" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      aws_cloudwatch_log_group.on_timer.arn,
      aws_cloudwatch_log_group.on_feed_msg.arn,
      aws_cloudwatch_log_group.on_post_msg.arn,
    ]
  }
}

resource "aws_iam_policy" "function_logging" {
  name        = "AllowLambdaLoggingPolicy"
  description = "Policy for Lambda CloudWatch logging"
  policy      = data.aws_iam_policy_document.allow_lambda_logging.json
}

resource "aws_iam_role_policy_attachment" "lambda_logging" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_logging.arn
}

data "aws_iam_policy_document" "allow_lambda_dynamodb" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:GetItem",
      "dynamodb:Scan",
      "dynamodb:BatchGetItem",
      "dynamodb:Query",
      "dynamodb:BatchWriteItem",
      "dynamodb:DeleteItem",
      "dynamodb:ConditionCheckItem",
    ]

    resources = [
      aws_dynamodb_table.email_sub.arn,
    ]
  }
}

resource "aws_iam_policy" "function_dynamodb" {
  name        = "AllowLambdaDynamoDBPolicy"
  description = "Policy for Lambda DynamoDB access"
  policy      = data.aws_iam_policy_document.allow_lambda_dynamodb.json
}

resource "aws_iam_role_policy_attachment" "lambda_dynamodb" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_dynamodb.arn
}

data "aws_iam_policy_document" "allow_lambda_sqs" {
  statement {
    effect = "Allow"
    actions = [
      "sqs:SendMessage",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
    ]

    resources = [
      aws_sqs_queue.email_sub_feeds.arn,
      aws_sqs_queue.email_sub_posts.arn,
    ]
  }
}

resource "aws_iam_policy" "function_sqs" {
  name        = "AllowLambdaSQSPolicy"
  description = "Policy for Lambda SQS access"
  policy      = data.aws_iam_policy_document.allow_lambda_sqs.json
}

resource "aws_iam_role_policy_attachment" "lambda_sqs" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_sqs.arn
}

data "aws_iam_policy_document" "allow_lambda_ses" {
  statement {
    effect = "Allow"
    actions = [
      "ses:SendEmail",
      "ses:SendRawEmail",
    ]

    resources = [
      "*",
    ]
  }
}

resource "aws_iam_policy" "function_ses" {
  name        = "AllowLambdaSESPolicy"
  description = "Policy for Lambda SES access"
  policy      = data.aws_iam_policy_document.allow_lambda_ses.json
}

resource "aws_iam_role_policy_attachment" "lambda_ses" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_ses.arn
}
