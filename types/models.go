// Package types exports stable string constants for common API literals (TYP-01).
package types

// Model name constants (any string is accepted by APIs).
// Chat flagship: Grok 4.5 (docs.x.ai). Prefer ModelGrok45Latest for “always newest 4.5”.
// Image/Video: Grok Imagine; video 1.5 is newer than base grok-imagine-video.
const (
	ModelGrok45                = "grok-4.5"
	ModelGrok45Latest          = "grok-4.5-latest"
	ModelGrok420               = "grok-4.20"
	ModelGrok3                 = "grok-3"
	ModelImagineImage          = "grok-imagine-image"
	ModelImagineImagePro       = "grok-imagine-image-pro"
	ModelImagineImageQuality   = "grok-imagine-image-quality" // tool / higher-quality path in docs
	ModelImagineVideo          = "grok-imagine-video"
	ModelImagineVideo15        = "grok-imagine-video-1.5"
	ModelImagineVideo15Preview = "grok-imagine-video-1.5-preview" // alias of 1.5 line
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

// ImageAspectRatio values (all accepted by image.WithAspectRatio).
const (
	Aspect1_1  = "1:1"
	Aspect16_9 = "16:9"
	Aspect9_16 = "9:16"
	Aspect3_4  = "3:4"
	Aspect4_3  = "4:3"
	Aspect2_3  = "2:3"
	Aspect3_2  = "3:2"
	AspectAuto = "auto"
)

// ImageResolution values for image.WithResolution.
const (
	ImageRes1K = "1k"
	ImageRes2K = "2k"
)

// VideoAspectRatio / resolution values (all accepted by video.WithAspectRatio
// / video.WithResolution).
const (
	VideoAspect1_1  = "1:1"
	VideoAspect16_9 = "16:9"
	VideoAspect9_16 = "9:16"
	VideoAspect4_3  = "4:3"
	VideoAspect3_4  = "3:4"
	VideoAspect3_2  = "3:2"
	VideoAspect2_3  = "2:3"
	VideoRes480p    = "480p"
	VideoRes720p    = "720p"
)

// Files content download formats for files.WithContentFormat.
const (
	ContentFormatOriginal = "original"
	ContentFormatText     = "text"
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
