package iec104

// Type identifiers (IEC 60870-5-101/104 Table 8).
const (
	// Monitoring – process information
	M_SP_NA_1 = 1  // single-point, no time
	M_DP_NA_1 = 3  // double-point, no time
	M_ME_NA_1 = 9  // normalized measured, no time
	M_ME_NB_1 = 11 // scaled measured, no time
	M_ME_NC_1 = 13 // short float, no time
	M_ME_ND_1 = 21 // normalized measured, no quality

	// Monitoring – with CP56Time2a
	M_SP_TB_1 = 30 // single-point, CP56Time2a
	M_DP_TB_1 = 31 // double-point, CP56Time2a
	M_ME_TA_1 = 34 // normalized measured, CP56Time2a
	M_ME_TB_1 = 35 // scaled measured, CP56Time2a
	M_ME_TF_1 = 36 // short float, CP56Time2a

	// Control
	C_SC_NA_1 = 45  // single command
	C_DC_NA_1 = 46  // double command
	C_SE_NA_1 = 48  // setpoint normalized
	C_SE_NB_1 = 49  // setpoint scaled
	C_SE_NC_1 = 50  // setpoint short float

	// System
	C_IC_NA_1 = 100 // general interrogation
	C_CI_NA_1 = 101 // counter interrogation
	C_CS_NA_1 = 103 // clock synchronisation
)

// Cause of transmission values.
const (
	CotPeriodic        = 1
	CotBackground      = 2
	CotSpontaneous     = 3
	CotInitialized     = 4
	CotRequest         = 5
	CotActivation      = 6
	CotActCon          = 7
	CotDeactivation    = 8
	CotDeactCon        = 9
	CotActTerm         = 10
	CotInterrogated    = 20
	CotUnknownTypeID   = 44
	CotUnknownCOT      = 45
	CotUnknownCA       = 46
	CotUnknownIOA      = 47
)

// IsDigitalType returns true for single/double-point TypeIDs.
func IsDigitalType(typeID int) bool {
	switch typeID {
	case M_SP_NA_1, M_SP_TB_1, M_DP_NA_1, M_DP_TB_1:
		return true
	}
	return false
}

// IsAnalogType returns true for measured value TypeIDs.
func IsAnalogType(typeID int) bool {
	switch typeID {
	case M_ME_NA_1, M_ME_NB_1, M_ME_NC_1, M_ME_ND_1,
		M_ME_TA_1, M_ME_TB_1, M_ME_TF_1:
		return true
	}
	return false
}

// TypeIDName returns a human-readable label.
func TypeIDName(id int) string {
	names := map[int]string{
		M_SP_NA_1: "M_SP_NA_1",
		M_DP_NA_1: "M_DP_NA_1",
		M_ME_NA_1: "M_ME_NA_1",
		M_ME_NB_1: "M_ME_NB_1",
		M_ME_NC_1: "M_ME_NC_1",
		M_ME_ND_1: "M_ME_ND_1",
		M_SP_TB_1: "M_SP_TB_1",
		M_DP_TB_1: "M_DP_TB_1",
		M_ME_TA_1: "M_ME_TA_1",
		M_ME_TB_1: "M_ME_TB_1",
		M_ME_TF_1: "M_ME_TF_1",
		C_SC_NA_1: "C_SC_NA_1",
		C_DC_NA_1: "C_DC_NA_1",
		C_SE_NA_1: "C_SE_NA_1",
		C_SE_NB_1: "C_SE_NB_1",
		C_SE_NC_1: "C_SE_NC_1",
		C_IC_NA_1: "C_IC_NA_1",
		C_CS_NA_1: "C_CS_NA_1",
	}
	if n, ok := names[id]; ok {
		return n
	}
	return "UNKNOWN"
}
