// Package types exports stable string constants for common API literals (TYP-01).
package types

// Model name constants (any string is accepted by APIs).
const (
	ModelGrok45                = "grok-4.5"
	ModelGrok45Latest          = "grok-4.5-latest"
	ModelGrok420               = "grok-4.20"
	ModelGrok3                 = "grok-3"
	ModelImagineImage          = "grok-imagine-image"
	ModelImagineImagePro       = "grok-imagine-image-pro"
	ModelImagineVideo          = "grok-imagine-video"
	ModelImagineVideo15Preview = "grok-imagine-video-1.5-preview"
)

// ReasoningEffort values.
const (
	ReasoningNone   = "none"
	ReasoningLow    = "low"
	ReasoningMedium = "medium"
	ReasoningHigh   = "high"
)

// ServiceTier values.
const (
	ServiceTierDefault  = "default"
	ServiceTierPriority = "priority"
)

// ImageFormat values.
const (
	ImageFormatURL    = "url"
	ImageFormatBase64 = "base64"
)

// ImageAspectRatio common values.
const (
	Aspect1_1  = "1:1"
	Aspect16_9 = "16:9"
	Aspect9_16 = "9:16"
	Aspect3_4  = "3:4"
	Aspect4_3  = "4:3"
	AspectAuto = "auto"
)

// VideoAspectRatio / resolution common values.
const (
	VideoAspect16_9 = "16:9"
	VideoAspect9_16 = "9:16"
	VideoRes480p    = "480p"
	VideoRes720p    = "720p"
)

// ToolChoice modes.
const (
	ToolModeAuto     = "auto"
	ToolModeNone     = "none"
	ToolModeRequired = "required"
)

// Search modes.
const (
	SearchModeAuto = "auto"
	SearchModeOn   = "on"
	SearchModeOff  = "off"
)

// ImageDetail values for chat image content.
const (
	ImageDetailAuto = "auto"
	ImageDetailLow  = "low"
	ImageDetailHigh = "high"
)
