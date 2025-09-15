package sdfl

type SeqType int

const (
	SEQ_TYPE_CALL SeqType = iota
	SEQ_TYPE_ARG
	SEQ_TYPE_VAL
	SEQ_TYPE_LIT
	SEQ_TYPE_LEFT
	SEQ_TYPE_RIGHT
)

type StackSeqObject struct {
	SeqType  SeqType
	RuleType *RuleType
	Id       *string
}

func parseSeq() {

}
