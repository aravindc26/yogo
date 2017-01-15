package yoga

type Direction int

const (
	DirectionInherit Direction = iota
	DirectionLTR
	DirectionRTL
)

type MeasureMode int

const (
	MeasureModeUndefined MeasureMode = iota
	MeasureModeExactly
	MeasureModeAtmost
)

type FlexDirection int

const (
	FlexDirectionColumn FlexDirection = iota
	FlexDirectionColumnReverse
	FlexDirectionRow
	FlexDirectionRowReverse
)

type Justify int

const (
	JustifyFlexStart Justify = iota
	JustifyCenter
	JustifyFlexEnd
	JustifySpaceBetween
	JustifySpaceAround
)

type Align int

const (
	AlignAuto Align = iota
	AlignFlexStart
	AlignCenter
	AlignFlexEnd
	AlignStretch
	AlignBaseLine
)

type PositionType int

const (
	PositionTypeRelative PositionType = iota
	PositionTypeAbsolute
)

type Wrap int

const (
	WrapNoWrap Wrap = iota
	WrapWrap
)

type Overflow int

const (
	OverflowVisible Overflow = iota
	OverflowHidden
	OverflowScroll
)

type Unit int

const (
	UnitUndefined Unit = iota
	UnitPixel
	UnitPercent
)

type Edge int

const (
	EdgeLeft Edge = iota
	EdgeTop
	EdgeRight
	EdgeBottom
	EdgeStart
	EdgeEnd
	EdgeHorizontal
	EdgeVertical
	EdgeAll
)
