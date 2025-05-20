resource "aws_dynamodb_table" "email_sub" {
  billing_mode = "PAY_PER_REQUEST"
  name         = "email-sub"
  hash_key     = "pk"
  range_key    = "sk"

  attribute {
    name = "pk"
    type = "S"
  }
  attribute {
    name = "sk"
    type = "S"
  }
}
