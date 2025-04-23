package resources

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/schema"
)

var IntegrationSlackBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeToken,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSigningSecret,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeWebhookUrl,
			Required: false,
		},
	},
}

var IntegrationEmailBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSmtpTls,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSmtpHost,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeSmtpPort,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSmtpsPort,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSmtpUsername,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSmtpPassword,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeFrom,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeTo,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeCc,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeBcc,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSubject,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeResponseUrl,
			Required: false,
		},
	},
}

var IntegrationTeamsBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeWebhookUrl,
			Required: true,
		},
	},
}

var TriggerScheduleBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSchedule,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePipeline,
			Required: true,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var TriggerIntervalBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSchedule,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePipeline,
			Required: true,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
}

var TriggerQueryBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			// Schedule is not a required attribute for Query Trigger, default to every 15 minutes
			Name: schema.AttributeTypeSchedule,
		},
		{
			Name:     schema.AttributeTypeSql,
			Required: true,
		},
		{
			Name: schema.AttributeTypePrimaryKey,
		},
		{
			Name:     schema.AttributeTypeDatabase,
			Required: false,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeCapture,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var TriggerHttpBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name: schema.AttributeTypePipeline,
		},
		{
			Name: schema.AttributeTypeExecutionMode,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeEnabled,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeMethod,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var PipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDocumentation,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypePipelineStep,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var PipelineOutputBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name:     schema.AttributeTypeValue,
			Required: true,
		},
	},
}

var PipelineParamBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeType,
		},
		{
			Name: schema.AttributeTypeDefault,
		},
		{
			Name:     schema.AttributeTypeEnum,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTags,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeOptional,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeFormat,
			Required: false,
		},
	},
}

var PipelineStepHttpBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeUrl,
			Required: true,
		},
		{
			Name: schema.AttributeTypeMethod,
		},
		{
			Name: schema.AttributeTypeCaCertPem,
		},
		{
			Name: schema.AttributeTypeInsecure,
		},
		{
			Name: schema.AttributeTypeRequestBody,
		},
		{
			Name: schema.AttributeTypeRequestHeaders,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypePipelineBasicAuth,
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineBasicAuthBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeUsername,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePassword,
			Required: true,
		},
	},
}

var PipelineStepSleepBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeDuration,
			Required: true,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepEmailBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeTo,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeFrom,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeSmtpUsername,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeSmtpPassword,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeHost,
			Required: true,
		},
		{
			Name:     schema.AttributeTypePort,
			Required: true,
		},
		{
			Name: schema.AttributeTypeSenderName,
		},
		{
			Name: schema.AttributeTypeCc,
		},
		{
			Name: schema.AttributeTypeBcc,
		},
		{
			Name: schema.AttributeTypeBody,
		},
		{
			Name: schema.AttributeTypeContentType,
		},
		{
			Name: schema.AttributeTypeSubject,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepQueryBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeSql,
		},
		{
			Name: schema.AttributeTypeDatabase,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepTransformBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeValue,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepPipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypePipeline,
		},
		{
			Name: schema.AttributeTypeArgs,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepFunctionBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeSource,
			Required: true,
		},
		{
			Name: schema.AttributeTypeHandler,
		},
		{
			Name: schema.AttributeTypeRuntime,
		},
		{
			Name: schema.AttributeTypeEnv,
		},
		{
			Name: schema.AttributeTypeEvent,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
		{
			Type: schema.BlockTypeLoop,
		},
	},
}

var PipelineStepContainerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeTimeout,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeImage,
			// Required: true,
		},
		{
			Name: schema.AttributeTypeSource,
			// Required: true,
		},
		{
			Name: schema.AttributeTypeCmd,
		},
		{
			Name: schema.AttributeTypeEnv,
		},
		{
			Name: schema.AttributeTypeEntrypoint,
		},
		{
			Name: schema.AttributeTypeCpuShares,
		},
		{
			Name: schema.AttributeTypeMemory,
		},
		{
			Name: schema.AttributeTypeMemoryReservation,
		},
		{
			Name: schema.AttributeTypeMemorySwap,
		},
		{
			Name: schema.AttributeTypeMemorySwappiness,
		},
		{
			Name: schema.AttributeTypeReadOnly,
		},
		{
			Name: schema.AttributeTypeUser,
		},
		{
			Name: schema.AttributeTypeWorkdir,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
	},
}

var PipelineStepInputBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name: schema.AttributeTypeOptions,
		},
		{
			Name: schema.AttributeTypePrompt,
		},
		{
			Name:     schema.AttributeTypeNotifier,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeType,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeTo,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeCc,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeBcc,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSubject,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeChannel,
			Required: false,
		},
		{
			Name: schema.AttributeTypeMaxConcurrency,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
		{
			Type:       schema.BlockTypeOption,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeLoop,
		},
	},
}

var PipelineStepMessageBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeTitle,
		},
		{
			Name: schema.AttributeTypeDescription,
		},
		{
			Name: schema.AttributeTypeForEach,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
		{
			Name: schema.AttributeTypeIf,
		},
		{
			Name:     schema.AttributeTypeNotifier,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeText,
			Required: true,
		},
		{
			Name:     schema.AttributeTypeTo,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeCc,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeBcc,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeSubject,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeChannel,
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeError,
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type: schema.BlockTypeRetry,
		},
		{
			Type: schema.BlockTypeThrow,
		},
		{
			Type: schema.BlockTypeLoop,
		},
	},
}

var PipelineStepInputNotifyBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: schema.AttributeTypeIntegration,
		},
		{
			Name: schema.AttributeTypeChannel,
		},
	},
}
