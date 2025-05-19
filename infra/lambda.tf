resource "aws_lambda_function" "on_request" {
  function_name = "on-request"
  description   = "Email-sub Lambda function for: on-request"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"
  filename      = "../dist/on-request.zip"
  runtime       = "provided.al2023"
  architectures = ["arm64"]
}
