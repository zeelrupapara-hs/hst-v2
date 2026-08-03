// Package model is the wire contract every service shares: subjects, events and trade envelopes.
//
// The rules below are the whole convention. A name you have not seen before should be derivable
// from them without asking anyone.
//
// 1. An enum is TypeName_lower_snake_value, and carries both a _name and a _value map.
//
//	OrderType_buy_stop_limit   OrderEvent_new_order   OrderTime_gtc
//
// 2. A request field uses its enum type, never int32. Validity is membership in _name, never a
// numeric bound, because a bound silently outlaws whichever value is added last.
//
// 3. A pointer means the caller did not supply the field. Everything else is a value.
//
//	PriceSL *float64   ExpiryAt *int64
//
// 4. A subject beginning websocket. reaches a client. A subject beginning system. never does.
// There is no third form, and neither carries a shard.
//
// 5. A function's verb says what it does: New creates, Cook is the per-tick pass, Calculate works
// out money, Check answers yes or no, Load hydrates at boot, SaveAndPublish writes then announces.
//
// 6. A comment is one line, or there is no comment.
package model
