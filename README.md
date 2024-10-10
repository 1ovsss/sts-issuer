# Sts-issuer

A service that issues temporary tokens for access to S3. The service takes the SA key and issues a temporary token with the same or reduced rights to the selected resources  

# Envs
`id` - the postfix of the parameters by which all the necessary settings are grouped into a json schema. In other words, the role. For example - `OssMlStage`

|  Name  | Description | Value |
| ------ |  -------    | ------|
|AWS_REGION | aws region |`ru-central1`|
|AWS_ACCESS_KEY_ID |   Access key root СА | `secret` |
|AWS_SECRET_ACCESS_KEY| Secret key root СА | `secret` |
|YC_STS_URL| STS endpoint | `https://sts.yandexcloud.net/` |
|STS_PORTS| Port for service in server mode | `3333` |
|STS_CRON_IDENTIFIER| the role ID in the cronjob operation mode, using this flag, the service will send a notification and shut down | `null` |
|RC_WEBHOOK| RocketChat Webhook | `secret` |
|RC_CHANNEL_`id`| RokceChat channel for notifications | `sts-issuer` |
|STS_EXPIRES_IN_`id`| The time after which the token will be revoked, in seconds (15 minutes-12 hours) | `900` |
|STS_ARN_`id`| The name of the arn role (which is not used by Yandex) | `arn:yc:iam::<folder>:role/<name>` |
|STS_POLICY_`id`_*| The description of the policy is used, separated by a postfix | `<policy>` |

# Api 
`/v1/list` - a list of all available roles and their settings   
`/vi/issue?id=<id>` - accepts the `id` of the role and returns secret/access/session tokens

# Policy
The policy consists of three parts and is described in a json schema

| Env | Description | Value |
| ------ | ------ | -----
| Effect | Allow or deny an action    | Allow/Deny
| Action    | Actions with a bucket  | [Actions](https://yandex.cloud/ru/docs/storage/s3/api-ref/policy/actions)
| Resource | The resources provided, the bucket. Can be described as `<bucket>` or `<bucket>/*`  | `arn:aws:s3:::<bucket>`

Policy example:  
`- name: STS_POLICY_ProdBucketApp_2  
  value: '{"Effect": "Allow", "Action": ["s3:ListBucket"], "Resource": ["arn:aws:s3:::s3-prod-bucket-app"]}'`