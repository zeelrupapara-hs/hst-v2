package model

import "testing"

func TestPermissionsFlagsNotifyBits(t *testing.T) {
	const base = PermissionsFlags_enable_connection | PermissionsFlags_risk_warning
	dealsOrders := PermissionsFlags_notify_deals | PermissionsFlags_notify_orders
	flags := base | dealsOrders

	if !flags.Has(PermissionsFlags_notify_deals) {
		t.Fatal("expected notify_deals")
	}
	if !flags.Has(PermissionsFlags_notify_orders) {
		t.Fatal("expected notify_orders")
	}
	if flags.Has(PermissionsFlags_notify_balances) {
		t.Fatal("did not expect notify_balances")
	}

	cleared := (flags &^ (PermissionsFlags_notify_deals | PermissionsFlags_notify_orders | PermissionsFlags_notify_balances)) |
		PermissionsFlags_notify_balances
	if cleared.Has(PermissionsFlags_notify_deals) || cleared.Has(PermissionsFlags_notify_orders) {
		t.Fatal("notify deals/orders should be cleared")
	}
	if !cleared.Has(PermissionsFlags_notify_balances) {
		t.Fatal("expected notify_balances after merge")
	}
	if !cleared.Has(PermissionsFlags_enable_connection) {
		t.Fatal("non-notify permission bits must be preserved")
	}
}

func TestPermissionsFlagsGroupDefaultIncludesAllNotify(t *testing.T) {
	want := PermissionsFlags_notify_deals | PermissionsFlags_notify_orders | PermissionsFlags_notify_balances
	if PermissionsFlags_group_default&want != want {
		t.Fatalf("group default %d missing notify bits %d", PermissionsFlags_group_default, want)
	}
}
