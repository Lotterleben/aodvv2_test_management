package aodvv2_test_management

/* route states */
const
(
    ROUTE_STATE_ACTIVE = iota
    ROUTE_STATE_IDLE = iota
    ROUTE_STATE_INVALID = iota
    ROUTE_STATE_TIMED = iota
)

const Test_string = "xoxotesttest"

const Tmpl_sent_rreq = "{\"log_type\": \"sent_rreq\", "+
                        "\"log_data\": {"+
                                "\"orig_addr\": \"{{.Orig_addr}}\", "+
                                "\"orig_seqnum\": {{.Orig_seqnum}}, "+
                                "\"targ_addr\": \"{{.Targ_addr}}\", "+
                                "\"metric\": {{.Metric}}}}"

const Tmpl_received_rreq = "{\"log_type\": \"received_rreq\", "+
                            "\"log_data\":{"+
                                "\"last_hop\": \"{{.Last_hop}}\", "+
                                "\"orig_addr\": \"{{.Orig_addr}}\", "+
                                "\"orig_seqnum\": {{.Orig_seqnum}}, "+
                                "\"targ_addr\": \"{{.Targ_addr}}\", "+
                                "\"metric\": {{.Metric}}}}"

const Tmpl_added_rt_entry = "{\"log_type\": \"added_rt_entry\", "+
                             "\"log_data\": {"+
                                "\"addr\": \"{{.Addr}}\", "+
                                "\"next_hop\": \"{{.Next_hop}}\", "+
                                "\"seqnum\": {{.Seqnum}}, "+
                                "\"metric\": {{.Metric}}, "+
                                "\"state\": {{.State}}}}"

const Tmpl_sent_rrep =  "{\"log_type\": \"sent_rrep\", "+
                        "\"log_data\": {"+
                                "\"next_hop\": \"{{.Next_hop}}\", "+
                                "\"orig_addr\": \"{{.Orig_addr}}\", "+
                                "\"orig_seqnum\": {{.Orig_seqnum}}, "+
                                "\"targ_addr\": \"{{.Targ_addr}}\"}}"

const Tmpl_received_rrep = "{\"log_type\": \"received_rrep\", "+
                            "\"log_data\":{"+
                                "\"last_hop\": \"{{.Last_hop}}\", "+
                                "\"orig_addr\": \"{{.Orig_addr}}\", "+
                                "\"orig_seqnum\": {{.Orig_seqnum}}, "+
                                "\"targ_addr\": \"{{.Targ_addr}}\", "+
                                "\"targ_seqnum\": {{.Targ_seqnum}}}}"
