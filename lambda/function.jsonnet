{
  FunctionName: 'tracer',
  MemorySize: 128,
  Handler: 'index.handler',
  Role: 'arn:aws:iam::{account_id}:role/{role_name}',
  Runtime: 'provided.al2',
}
