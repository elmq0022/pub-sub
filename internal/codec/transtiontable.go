package codec

type STATE uint8

const (
	ST_ERROR STATE = iota
	ST_START

	ST_CR_END
	ST_DONE

	// CONNECT {}\r\n
	ST_CMD_C
	ST_CMD_CO
	ST_CMD_CON
	ST_CMD_CONN
	ST_CMD_CONNE
	ST_CMD_CONNEC
	ST_CMD_CONNECT
	ST_CONNECT_SPACE
	ST_CONNECT_LBRACE
	ST_CONNECT_RBRACE

	// PING\r\n
	ST_CMD_P
	ST_CMD_PI
	ST_CMD_PIN
	ST_CMD_PING

	// PONG\r\n
	// ST_CMD_P
	ST_CMD_PO
	ST_CMD_PON
	ST_CMD_PONG

	// SUB <subject> <sid>\r\n
	ST_CMD_S
	ST_CMD_SU
	ST_CMD_SUB
	ST_SUB_SPACE
	ST_SUB_SUBJECT
	ST_SUB_SUBJECT_SPACE

	ST_SUB_SUBJECT_DOT
	ST_SUB_SUBJECT_STAR
	ST_SUB_SUBJECT_GT

	ST_SUB_SID

	// PUB <subject> <#bytes>\r\n[payload]\r\n
	// ST_CMD_P
	ST_CMD_PU
	ST_CMD_PUB
	ST_PUB_SPACE
	ST_PUB_SUBJECT
	ST_PUB_SUBJECT_SPACE
	ST_PUB_SUBJECT_DOT
	ST_PUB_NUM_BYTES
	ST_PUB_CR
	ST_PUB_PAYLOAD

	// UNSUB <sid>\r\n
	ST_CMD_U
	ST_CMD_UN
	ST_CMD_UNS
	ST_CMD_UNSU
	ST_CMD_UNSUB
	ST_UNSUB_SPACE
	ST_UNSUB_SID
)

var transitionTable = buildTransitionTable()

const nStates = int(ST_UNSUB_SID) + 1

var digits = []byte("0123456789")
var alphas = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

func buildTransitionTable() [nStates][256]STATE {
	var t [nStates][256]STATE
	t[ST_START]['C'] = ST_CMD_C
	t[ST_CMD_C]['O'] = ST_CMD_CO
	t[ST_CMD_CO]['N'] = ST_CMD_CON
	t[ST_CMD_CON]['N'] = ST_CMD_CONN
	t[ST_CMD_CONN]['E'] = ST_CMD_CONNE
	t[ST_CMD_CONNE]['C'] = ST_CMD_CONNEC
	t[ST_CMD_CONNEC]['T'] = ST_CMD_CONNECT
	t[ST_CMD_CONNECT][' '] = ST_CONNECT_SPACE
	t[ST_CONNECT_SPACE]['{'] = ST_CONNECT_LBRACE
	t[ST_CONNECT_LBRACE]['}'] = ST_CONNECT_RBRACE
	t[ST_CONNECT_RBRACE]['\r'] = ST_CR_END
	t[ST_CR_END]['\n'] = ST_DONE

	t[ST_START]['P'] = ST_CMD_P
	t[ST_CMD_P]['I'] = ST_CMD_PI
	t[ST_CMD_PI]['N'] = ST_CMD_PIN
	t[ST_CMD_PIN]['G'] = ST_CMD_PING
	t[ST_CMD_PING]['\r'] = ST_CR_END

	t[ST_CMD_P]['O'] = ST_CMD_PO
	t[ST_CMD_PO]['N'] = ST_CMD_PON
	t[ST_CMD_PON]['G'] = ST_CMD_PONG
	t[ST_CMD_PONG]['\r'] = ST_CR_END

	// SUB <subject> <sid>\r\n
	t[ST_START]['S'] = ST_CMD_S
	t[ST_CMD_S]['U'] = ST_CMD_SU
	t[ST_CMD_SU]['B'] = ST_CMD_SUB
	t[ST_CMD_SUB][' '] = ST_SUB_SPACE

	// from a space to the subject
	for _, c := range alphas {
		t[ST_SUB_SPACE][c] = ST_SUB_SUBJECT
	}
	for _, c := range digits {
		t[ST_SUB_SPACE][c] = ST_SUB_SUBJECT
	}
	// subjects can start with a wildcard
	t[ST_SUB_SPACE]['*'] = ST_SUB_SUBJECT_STAR
	t[ST_SUB_SPACE]['>'] = ST_SUB_SUBJECT_GT

	// through valid subject chars
	for _, c := range alphas {
		t[ST_SUB_SUBJECT][c] = ST_SUB_SUBJECT
	}
	for _, c := range digits {
		t[ST_SUB_SUBJECT][c] = ST_SUB_SUBJECT
	}
	t[ST_SUB_SUBJECT][' '] = ST_SUB_SUBJECT_SPACE

	// dot can move to a subject or an * or a >
	t[ST_SUB_SUBJECT]['.'] = ST_SUB_SUBJECT_DOT
	for _, c := range alphas {
		t[ST_SUB_SUBJECT_DOT][c] = ST_SUB_SUBJECT
	}
	for _, c := range digits {
		t[ST_SUB_SUBJECT_DOT][c] = ST_SUB_SUBJECT
	}
	t[ST_SUB_SUBJECT_DOT]['*'] = ST_SUB_SUBJECT_STAR
	t[ST_SUB_SUBJECT_DOT]['>'] = ST_SUB_SUBJECT_GT

	// subject star must go back to do or end the subject
	t[ST_SUB_SUBJECT_STAR]['.'] = ST_SUB_SUBJECT_DOT
	t[ST_SUB_SUBJECT_STAR][' '] = ST_SUB_SUBJECT_SPACE

	// a > must end a subject
	t[ST_SUB_SUBJECT_GT][' '] = ST_SUB_SUBJECT_SPACE

	for _, d := range digits {
		t[ST_SUB_SUBJECT_SPACE][d] = ST_SUB_SID
	}

	for _, d := range digits {
		t[ST_SUB_SID][d] = ST_SUB_SID
	}
	t[ST_SUB_SID]['\r'] = ST_CR_END

	// PUB <subject> <#bytes>\r\n[payload]\r\n
	// ST_CMD_P
	t[ST_CMD_P]['U'] = ST_CMD_PU
	t[ST_CMD_PU]['B'] = ST_CMD_PUB
	t[ST_CMD_PUB][' '] = ST_PUB_SPACE

	for _, c := range alphas {
		t[ST_PUB_SPACE][c] = ST_PUB_SUBJECT
	}
	for _, c := range digits {
		t[ST_PUB_SPACE][c] = ST_PUB_SUBJECT
	}

	for _, c := range alphas {
		t[ST_PUB_SUBJECT][c] = ST_PUB_SUBJECT
		t[ST_PUB_SUBJECT_DOT][c] = ST_PUB_SUBJECT
	}
	for _, c := range digits {
		t[ST_PUB_SUBJECT][c] = ST_PUB_SUBJECT
		t[ST_PUB_SUBJECT_DOT][c] = ST_PUB_SUBJECT
	}
	t[ST_PUB_SUBJECT]['.'] = ST_PUB_SUBJECT_DOT
	t[ST_PUB_SUBJECT][' '] = ST_PUB_SUBJECT_SPACE

	for _, c := range digits {
		t[ST_PUB_SUBJECT_SPACE][c] = ST_PUB_NUM_BYTES
		t[ST_PUB_NUM_BYTES][c] = ST_PUB_NUM_BYTES
	}

	// Will read the n bytes after parsing the
	// first line of the publish command and finish
	// the command that way as the payload can't be
	// reliably parsed otherwise.
	// TODO: make sure to validate CRLF for payload
	t[ST_PUB_NUM_BYTES]['\r'] = ST_PUB_CR
	t[ST_PUB_CR]['\n'] = ST_PUB_PAYLOAD

	return t
}
