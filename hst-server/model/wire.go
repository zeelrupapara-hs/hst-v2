package model

import wire "hstmodel"

// The wire contract, defined once in the shared model and re-exported here so this package stays
// the single import for a handler. Changing any of these means changing hst-model.

type (
	OrderType       = wire.OrderType
	OrderState      = wire.OrderState
	OrderFilling    = wire.OrderFilling
	OrderTime       = wire.OrderTime
	OrderReason     = wire.OrderReason
	OrderActivation = wire.OrderActivation
	ActivationFlags = wire.ActivationFlags
	ModifyFlags     = wire.ModifyFlags

	EventType         = wire.EventType
	OrderEventType    = wire.OrderEventType
	PositionEventType = wire.PositionEventType
	DealingEventType  = wire.DealingEventType

	TradeRequest  = wire.TradeRequest
	TradeResult   = wire.TradeResult
	OrderEvent    = wire.OrderEvent
	PositionEvent = wire.PositionEvent
	DealingEvent  = wire.DealingEvent

	PositionAction = wire.PositionAction
	DealAction     = wire.DealAction
	DealEntry      = wire.DealEntry
)

const (
	OrderType_buy             = wire.OrderType_buy
	OrderType_sell            = wire.OrderType_sell
	OrderType_buy_limit       = wire.OrderType_buy_limit
	OrderType_sell_limit      = wire.OrderType_sell_limit
	OrderType_buy_stop        = wire.OrderType_buy_stop
	OrderType_sell_stop       = wire.OrderType_sell_stop
	OrderType_buy_stop_limit  = wire.OrderType_buy_stop_limit
	OrderType_sell_stop_limit = wire.OrderType_sell_stop_limit
	OrderType_close_by        = wire.OrderType_close_by

	OrderState_started        = wire.OrderState_started
	OrderState_placed         = wire.OrderState_placed
	OrderState_canceled       = wire.OrderState_canceled
	OrderState_partial        = wire.OrderState_partial
	OrderState_filled         = wire.OrderState_filled
	OrderState_rejected       = wire.OrderState_rejected
	OrderState_expired        = wire.OrderState_expired
	OrderState_request_add    = wire.OrderState_request_add
	OrderState_request_modify = wire.OrderState_request_modify
	OrderState_request_cancel = wire.OrderState_request_cancel

	OrderFilling_fok    = wire.OrderFilling_fok
	OrderFilling_ioc    = wire.OrderFilling_ioc
	OrderFilling_return = wire.OrderFilling_return
	OrderFilling_boc    = wire.OrderFilling_boc

	OrderTime_gtc           = wire.OrderTime_gtc
	OrderTime_day           = wire.OrderTime_day
	OrderTime_specified     = wire.OrderTime_specified
	OrderTime_specified_day = wire.OrderTime_specified_day

	OrderReason_client           = wire.OrderReason_client
	OrderReason_expert           = wire.OrderReason_expert
	OrderReason_dealer           = wire.OrderReason_dealer
	OrderReason_sl               = wire.OrderReason_sl
	OrderReason_tp               = wire.OrderReason_tp
	OrderReason_so               = wire.OrderReason_so
	OrderReason_rollover         = wire.OrderReason_rollover
	OrderReason_external_client  = wire.OrderReason_external_client
	OrderReason_vmargin          = wire.OrderReason_vmargin
	OrderReason_gateway          = wire.OrderReason_gateway
	OrderReason_signal           = wire.OrderReason_signal
	OrderReason_settlement       = wire.OrderReason_settlement
	OrderReason_transfer         = wire.OrderReason_transfer
	OrderReason_sync             = wire.OrderReason_sync
	OrderReason_external_service = wire.OrderReason_external_service
	OrderReason_migration        = wire.OrderReason_migration
	OrderReason_mobile           = wire.OrderReason_mobile
	OrderReason_web              = wire.OrderReason_web
	OrderReason_split            = wire.OrderReason_split
	OrderReason_corporate_action = wire.OrderReason_corporate_action
	OrderReason_ultency          = wire.OrderReason_ultency
	OrderReason_coverage         = wire.OrderReason_coverage

	OrderActivation_none      = wire.OrderActivation_none
	OrderActivation_pending   = wire.OrderActivation_pending
	OrderActivation_stoplimit = wire.OrderActivation_stoplimit
	OrderActivation_sl        = wire.OrderActivation_sl
	OrderActivation_tp        = wire.OrderActivation_tp
	OrderActivation_stopout   = wire.OrderActivation_stopout

	ActivationFlags_none      = wire.ActivationFlags_none
	ActivationFlags_no_limit  = wire.ActivationFlags_no_limit
	ActivationFlags_no_stop   = wire.ActivationFlags_no_stop
	ActivationFlags_no_slimit = wire.ActivationFlags_no_slimit
	ActivationFlags_no_sl     = wire.ActivationFlags_no_sl
	ActivationFlags_no_tp     = wire.ActivationFlags_no_tp
	ActivationFlags_no_so     = wire.ActivationFlags_no_so
	ActivationFlags_no_expiry = wire.ActivationFlags_no_expiry

	ModifyFlags_none        = wire.ModifyFlags_none
	ModifyFlags_admin       = wire.ModifyFlags_admin
	ModifyFlags_manager     = wire.ModifyFlags_manager
	ModifyFlags_restore     = wire.ModifyFlags_restore
	ModifyFlags_api_admin   = wire.ModifyFlags_api_admin
	ModifyFlags_api_manager = wire.ModifyFlags_api_manager
	ModifyFlags_api_server  = wire.ModifyFlags_api_server
	ModifyFlags_api_gateway = wire.ModifyFlags_api_gateway

	OrderEvent_new_order    = wire.OrderEvent_new_order
	OrderEvent_update_order = wire.OrderEvent_update_order
	OrderEvent_cancel_order = wire.OrderEvent_cancel_order

	PositionEvent_update   = wire.PositionEvent_update
	PositionEvent_close    = wire.PositionEvent_close
	PositionEvent_close_by = wire.PositionEvent_close_by

	DealingEvent_offer   = wire.DealingEvent_offer
	DealingEvent_confirm = wire.DealingEvent_confirm
	DealingEvent_requote = wire.DealingEvent_requote
	DealingEvent_reject  = wire.DealingEvent_reject
	DealingEvent_cancel  = wire.DealingEvent_cancel
	DealingEvent_accept  = wire.DealingEvent_accept
	DealingEvent_return  = wire.DealingEvent_return

	ShardCount = wire.ShardCount

	PositionAction_buy  = wire.PositionAction_buy
	PositionAction_sell = wire.PositionAction_sell

	DealAction_buy                = wire.DealAction_buy
	DealAction_sell               = wire.DealAction_sell
	DealAction_balance            = wire.DealAction_balance
	DealAction_credit             = wire.DealAction_credit
	DealAction_charge             = wire.DealAction_charge
	DealAction_correction         = wire.DealAction_correction
	DealAction_bonus              = wire.DealAction_bonus
	DealAction_commission         = wire.DealAction_commission
	DealAction_commission_daily   = wire.DealAction_commission_daily
	DealAction_commission_monthly = wire.DealAction_commission_monthly
	DealAction_agent_daily        = wire.DealAction_agent_daily
	DealAction_agent_monthly      = wire.DealAction_agent_monthly
	DealAction_interest           = wire.DealAction_interest
	DealAction_buy_canceled       = wire.DealAction_buy_canceled
	DealAction_sell_canceled      = wire.DealAction_sell_canceled
	DealAction_dividend           = wire.DealAction_dividend
	DealAction_dividend_franked   = wire.DealAction_dividend_franked
	DealAction_tax                = wire.DealAction_tax
	DealAction_agent              = wire.DealAction_agent
	DealAction_so_compensation    = wire.DealAction_so_compensation
	DealAction_so_compensation_cr = wire.DealAction_so_compensation_cr

	DealEntry_in     = wire.DealEntry_in
	DealEntry_out    = wire.DealEntry_out
	DealEntry_inout  = wire.DealEntry_inout
	DealEntry_out_by = wire.DealEntry_out_by
)

var (
	OrderType_name       = wire.OrderType_name
	OrderType_value      = wire.OrderType_value
	OrderState_name      = wire.OrderState_name
	OrderState_value     = wire.OrderState_value
	OrderFilling_name    = wire.OrderFilling_name
	OrderFilling_value   = wire.OrderFilling_value
	OrderTime_name       = wire.OrderTime_name
	OrderTime_value      = wire.OrderTime_value
	OrderReason_name     = wire.OrderReason_name
	OrderReason_value    = wire.OrderReason_value
	OrderActivation_name = wire.OrderActivation_name
	OrderEvent_name      = wire.OrderEvent_name
	PositionEvent_name   = wire.PositionEvent_name
	DealingEvent_name    = wire.DealingEvent_name

	ShardOf = wire.ShardOf

	PositionAction_name = wire.PositionAction_name
	DealAction_name     = wire.DealAction_name
	DealEntry_name      = wire.DealEntry_name
)
