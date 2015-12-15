# What this is
A tool to run automated tests on a network of [RIOT](https://github.com/RIOT-OS/RIOT)s which lets you invoke an action and then check if the code behaves as expected.
To do this, it uses [desvirt](https://github.com/Lotterleben/desvirt/tree/line_fix_2) to spawn a line of interconnected RIOT instances, which can now be triggered and queried through the functions provided by `management.go`. 

note that:
- I've originally written this to test my RIOT-AODVv2 code, but it can be used for other purposes as well. However, some paths are hard-coded because I was pressed for time, so if you actually intend to use this, drop me a line and I'll be happy to make it more accessible.
- I'm using a desvirt fork because the original version doesn't properly support lines and I've never gotten around to open a PR (note to self: to that.)

# Usage
## Prerequisites
TODO

## Writing tests
The main idea is that you trigger some kind of action in one (or more) RIOTs and check if the other RIOTs respond as expected. The former is accomplished by sending a valid command to a RIOT (If the default set of shell commands don't do what you want your RIOT to do, you'll have to write your own):

```c
/* run our custom send command on the first RIOT of the line */
riot_line := mgmt.Create_clean_setup("my_first_test")
riot_line[0].Channels.Send(fmt.Sprintf("send %s %s\n", end.Ip, mgmt.Test_string))
```

Now, we can query all RIOTs in the line about the data they've received, processed and sent. More specifically, we can specify what we *expect* them to have received/output/sent, yielding an error if this is not the case:

```c
/* Discover route at node 0...  */
riot_line[0].Channels.Expect_JSON(mgmt.Make_JSON_str(mgmt.Tmpl_sent_rreq, map[string]string{
    "Orig_addr": beginning.Ip,
    "Targ_addr": end.Ip,
    "Orig_seqnum": "1",
    "Metric": "0",
}))

/* check node 1 */
riot_line[1].Channels.Expect_JSON(mgmt.Make_JSON_str(mgmt.Tmpl_received_rreq, map[string]string{
    "Last_hop": beginning.Ip,
    "Orig_addr": beginning.Ip,
    "Targ_addr": end.Ip,
    "Orig_seqnum": "1",
    "Metric": "0",
}))
riot_line[1].Channels.Expect_JSON(mgmt.Make_JSON_str(mgmt.Tmpl_added_rt_entry, map[string]string{
    "Addr": beginning.Ip,
    "Next_hop": beginning.Ip,
    "Seqnum": "1",
    "Metric": "1",
    "State": strconv.Itoa(mgmt.ROUTE_STATE_ACTIVE),
}))
riot_line[1].Channels.Expect_JSON(mgmt.Make_JSON_str(mgmt.Tmpl_sent_rreq, map[string]string{
    "Orig_addr": beginning.Ip,
    "Targ_addr": end.Ip,
    "Orig_seqnum": "1",
    "Metric": "1",
}))

/* check node 2... */
```

(Note that the code you're trying to test has to output the information you're looking for as a JSON in order for this to work.)

## Running tests
TODO