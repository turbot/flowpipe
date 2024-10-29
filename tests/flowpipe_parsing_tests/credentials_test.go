package pipeline_test

// func TestCredentials(t *testing.T) {
// 	assert := assert.New(t)

// 	mod, err := load_mod.LoadPipelinesReturningItsMod(context.TODO(), "./pipelines/credentials.fp")
// 	assert.Nil(err)
// 	assert.NotNil(mod)
// 	if mod == nil {
// 		return
// 	}

// 	credential := mod.ResourceMaps.credentials["local.credential.aws.aws_static"]
// 	if credential == nil {
// 		assert.Fail("Credential not found")
// 		return
// 	}

// 	assert.Equal("credential.aws.aws_static", credential.Name())
// 	assert.Equal("aws", credential.GetCredentialType())

// 	awsCred := credential.(*modconfig.AwsCredential)
// 	assert.Equal("ASIAQGDFAKEKGUI5MCEU", *awsCred.AccessKey)
// }
