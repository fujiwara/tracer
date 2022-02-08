resource "aws_iam_role" "tracer" {
  name = "tracer"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "tracer" {
  name   = "tracer"
  path   = "/"
  policy = data.aws_iam_policy_document.tracer.json
}

resource "aws_iam_role_policy_attachment" "tracer" {
  role       = aws_iam_role.tracer.name
  policy_arn = aws_iam_policy.tracer.arn
}

data "aws_iam_policy_document" "tracer" {
  statement {
    actions = [
      "ecs:Describe*",
      "ecs:List*",
    ]
    resources = ["*"]
  }
  statement {
    actions = [
      "logs:GetLog*",
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["*"]
  }
}

resource "aws_cloudwatch_event_rule" "tracer" {
  name       = "tracer"
  is_enabled = true
  event_pattern = jsonencode({
    "detail-type" = [
      "ECS Task State Change"
    ]
    "source" = [
      "aws.ecs"
    ]
  })
}

resource "aws_cloudwatch_event_target" "tracer-lambda" {
  rule = aws_cloudwatch_event_rule.tracer.name
  arn  = data.aws_lambda_function.tracer.arn
}

data "aws_lambda_function" "tracer" {
  function_name = "tracer"
}
