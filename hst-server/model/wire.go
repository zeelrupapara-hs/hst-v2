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

type (
	CalcMode             = wire.CalcMode
	TradeMode            = wire.TradeMode
	ExecMode             = wire.ExecMode
	MarginMode           = wire.MarginMode
	StopOutMode          = wire.StopOutMode
	FreeMarginMode       = wire.FreeMarginMode
	FreeMarginProfitMode = wire.FreeMarginProfitMode
	TypeFlags            = wire.TypeFlags
	RouteFlags           = wire.RouteFlags
	RouteAction          = wire.RouteAction
	RouteCondition       = wire.RouteCondition
	QueryWhat            = wire.QueryWhat
	ConditionRule        = wire.ConditionRule
)

const (
	CalcMode_forex                           = wire.CalcMode_forex
	CalcMode_futures                         = wire.CalcMode_futures
	CalcMode_cfd                             = wire.CalcMode_cfd
	CalcMode_cfd_index                       = wire.CalcMode_cfd_index
	CalcMode_cfd_leverage                    = wire.CalcMode_cfd_leverage
	CalcMode_forex_no_leverage               = wire.CalcMode_forex_no_leverage
	CalcMode_exch_stocks                     = wire.CalcMode_exch_stocks
	CalcMode_exch_futures                    = wire.CalcMode_exch_futures
	CalcMode_exch_forts                      = wire.CalcMode_exch_forts
	CalcMode_exch_options                    = wire.CalcMode_exch_options
	CalcMode_exch_options_margin             = wire.CalcMode_exch_options_margin
	TradeMode_disabled                       = wire.TradeMode_disabled
	TradeMode_long_only                      = wire.TradeMode_long_only
	TradeMode_short_only                     = wire.TradeMode_short_only
	TradeMode_close_only                     = wire.TradeMode_close_only
	TradeMode_full                           = wire.TradeMode_full
	ExecMode_request                         = wire.ExecMode_request
	ExecMode_instant                         = wire.ExecMode_instant
	ExecMode_market                          = wire.ExecMode_market
	ExecMode_exchange                        = wire.ExecMode_exchange
	MarginMode_retail_netting                = wire.MarginMode_retail_netting
	MarginMode_exchange                      = wire.MarginMode_exchange
	MarginMode_retail_hedging                = wire.MarginMode_retail_hedging
	StopOutMode_percent                      = wire.StopOutMode_percent
	StopOutMode_money                        = wire.StopOutMode_money
	FreeMarginMode_not_use_pl                = wire.FreeMarginMode_not_use_pl
	FreeMarginMode_use_pl                    = wire.FreeMarginMode_use_pl
	FreeMarginMode_profit                    = wire.FreeMarginMode_profit
	FreeMarginMode_loss                      = wire.FreeMarginMode_loss
	FreeMarginProfitMode_day_profit_and_loss = wire.FreeMarginProfitMode_day_profit_and_loss
	FreeMarginProfitMode_day_profit_loss     = wire.FreeMarginProfitMode_day_profit_loss
	TypeFlags_none                           = wire.TypeFlags_none
	TypeFlags_buy                            = wire.TypeFlags_buy
	TypeFlags_sell                           = wire.TypeFlags_sell
	TypeFlags_buy_limit                      = wire.TypeFlags_buy_limit
	TypeFlags_sell_limit                     = wire.TypeFlags_sell_limit
	TypeFlags_buy_stop                       = wire.TypeFlags_buy_stop
	TypeFlags_sell_stop                      = wire.TypeFlags_sell_stop
	TypeFlags_buy_stop_limit                 = wire.TypeFlags_buy_stop_limit
	TypeFlags_sell_stop_limit                = wire.TypeFlags_sell_stop_limit
	RouteFlags_none                          = wire.RouteFlags_none
	RouteFlags_price                         = wire.RouteFlags_price
	RouteFlags_request                       = wire.RouteFlags_request
	RouteFlags_instant                       = wire.RouteFlags_instant
	RouteFlags_market                        = wire.RouteFlags_market
	RouteFlags_exchange                      = wire.RouteFlags_exchange
	RouteFlags_pending                       = wire.RouteFlags_pending
	RouteFlags_sltp                          = wire.RouteFlags_sltp
	RouteFlags_modify                        = wire.RouteFlags_modify
	RouteFlags_remove                        = wire.RouteFlags_remove
	RouteFlags_activate                      = wire.RouteFlags_activate
	RouteFlags_stop_limit                    = wire.RouteFlags_stop_limit
	RouteFlags_sl                            = wire.RouteFlags_sl
	RouteFlags_tp                            = wire.RouteFlags_tp
	RouteFlags_stop_out_order                = wire.RouteFlags_stop_out_order
	RouteFlags_stop_out_position             = wire.RouteFlags_stop_out_position
	RouteFlags_expiration                    = wire.RouteFlags_expiration
	RouteFlags_close_by                      = wire.RouteFlags_close_by
	RouteAction_delay_time                   = wire.RouteAction_delay_time
	RouteAction_delay_tick                   = wire.RouteAction_delay_tick
	RouteAction_clear_tp                     = wire.RouteAction_clear_tp
	RouteAction_clear_sl                     = wire.RouteAction_clear_sl
	RouteAction_clear_sltp                   = wire.RouteAction_clear_sltp
	RouteAction_dealer                       = wire.RouteAction_dealer
	RouteAction_dealer_online                = wire.RouteAction_dealer_online
	RouteAction_reject                       = wire.RouteAction_reject
	RouteAction_requote                      = wire.RouteAction_requote
	RouteAction_confirm_client               = wire.RouteAction_confirm_client
	RouteAction_confirm_market               = wire.RouteAction_confirm_market
	RouteAction_cancel_order                 = wire.RouteAction_cancel_order
	RouteCondition_datetime                  = wire.RouteCondition_datetime
	RouteCondition_symbol                    = wire.RouteCondition_symbol
	RouteCondition_volume                    = wire.RouteCondition_volume
	RouteCondition_deviation                 = wire.RouteCondition_deviation
	RouteCondition_time                      = wire.RouteCondition_time
	RouteCondition_weekday                   = wire.RouteCondition_weekday
	RouteCondition_comment                   = wire.RouteCondition_comment
	RouteCondition_expert                    = wire.RouteCondition_expert
	RouteCondition_signal                    = wire.RouteCondition_signal
	RouteCondition_dealer_login              = wire.RouteCondition_dealer_login
	RouteCondition_source_login              = wire.RouteCondition_source_login
	RouteCondition_deviation_spread          = wire.RouteCondition_deviation_spread
	RouteCondition_gap                       = wire.RouteCondition_gap
	RouteCondition_reason                    = wire.RouteCondition_reason
	RouteCondition_request_price             = wire.RouteCondition_request_price
	RouteCondition_value                     = wire.RouteCondition_value
	RouteCondition_current_spread            = wire.RouteCondition_current_spread
	RouteCondition_login                     = wire.RouteCondition_login
	RouteCondition_group                     = wire.RouteCondition_group
	RouteCondition_country                   = wire.RouteCondition_country
	RouteCondition_city                      = wire.RouteCondition_city
	RouteCondition_color                     = wire.RouteCondition_color
	RouteCondition_leverage                  = wire.RouteCondition_leverage
	RouteCondition_comment2                  = wire.RouteCondition_comment2
	RouteCondition_zip                       = wire.RouteCondition_zip
	RouteCondition_status                    = wire.RouteCondition_status
	RouteCondition_client_id                 = wire.RouteCondition_client_id
	RouteCondition_party_id                  = wire.RouteCondition_party_id
	RouteCondition_margin                    = wire.RouteCondition_margin
	RouteCondition_margin_level              = wire.RouteCondition_margin_level
	RouteCondition_margin_free               = wire.RouteCondition_margin_free
	RouteCondition_equity                    = wire.RouteCondition_equity
	RouteCondition_balance                   = wire.RouteCondition_balance
	RouteCondition_profit                    = wire.RouteCondition_profit
	RouteCondition_daily_deals               = wire.RouteCondition_daily_deals
	RouteCondition_daily_deals_period        = wire.RouteCondition_daily_deals_period
	RouteCondition_daily_profit              = wire.RouteCondition_daily_profit
	RouteCondition_position_volume           = wire.RouteCondition_position_volume
	RouteCondition_position_profit           = wire.RouteCondition_position_profit
	RouteCondition_position_age              = wire.RouteCondition_position_age
	RouteCondition_position_modify_time      = wire.RouteCondition_position_modify_time
	RouteCondition_position_average_time     = wire.RouteCondition_position_average_time
	RouteCondition_position_total            = wire.RouteCondition_position_total
	RouteCondition_position_total_symbol     = wire.RouteCondition_position_total_symbol
	RouteCondition_order_total               = wire.RouteCondition_order_total
	RouteCondition_order_total_symbol        = wire.RouteCondition_order_total_symbol
	RouteCondition_position_sl_touched       = wire.RouteCondition_position_sl_touched
	RouteCondition_position_tp_touched       = wire.RouteCondition_position_tp_touched
	RouteCondition_order_sl_touched          = wire.RouteCondition_order_sl_touched
	RouteCondition_order_tp_touched          = wire.RouteCondition_order_tp_touched
	RouteCondition_position_value            = wire.RouteCondition_position_value
	RouteCondition_order_in                  = wire.RouteCondition_order_in
	RouteCondition_order_out                 = wire.RouteCondition_order_out
	QueryWhat_account                        = wire.QueryWhat_account
	QueryWhat_positions                      = wire.QueryWhat_positions
	QueryWhat_orders                         = wire.QueryWhat_orders
	QueryWhat_state                          = wire.QueryWhat_state
	QueryWhat_symbols                        = wire.QueryWhat_symbols
	ConditionRule_equal                      = wire.ConditionRule_equal
	ConditionRule_not_equal                  = wire.ConditionRule_not_equal
	ConditionRule_greater                    = wire.ConditionRule_greater
	ConditionRule_not_less                   = wire.ConditionRule_not_less
	ConditionRule_less                       = wire.ConditionRule_less
	ConditionRule_not_greater                = wire.ConditionRule_not_greater
)

const (
	CalcMode_exch_bonds            = wire.CalcMode_exch_bonds
	CalcMode_serv_collateral       = wire.CalcMode_serv_collateral
	RouteFlags_dealer_pos_execute  = wire.RouteFlags_dealer_pos_execute
	RouteFlags_dealer_ord_pending  = wire.RouteFlags_dealer_ord_pending
	RouteFlags_dealer_pos_modify   = wire.RouteFlags_dealer_pos_modify
	RouteFlags_dealer_ord_modify   = wire.RouteFlags_dealer_ord_modify
	RouteFlags_dealer_ord_remove   = wire.RouteFlags_dealer_ord_remove
	RouteFlags_dealer_ord_activate = wire.RouteFlags_dealer_ord_activate
	RouteFlags_dealer_ord_slimit   = wire.RouteFlags_dealer_ord_slimit
	RouteFlags_dealer_close_by     = wire.RouteFlags_dealer_close_by
)
