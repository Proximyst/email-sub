resource "aws_sqs_queue" "email_sub_feeds" {
  name                       = "email-sub-feeds"
  delay_seconds              = 0
  max_message_size           = 16384
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 0
  visibility_timeout_seconds = 30
}

resource "aws_sqs_queue" "email_sub_posts" {
  name                       = "email-sub-posts"
  delay_seconds              = 0
  max_message_size           = 16384
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 0
  visibility_timeout_seconds = 30
}
