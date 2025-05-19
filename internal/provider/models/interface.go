package models

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type SpanningTree struct {
	Portfast  types.String `tfsdk:"portfast"`
	BpduGuard types.Bool   `tfsdk:"bpdu_guard"`
}

var DefaultSpanningTree = SpanningTree{
	Portfast:  types.StringValue(""),
	BpduGuard: types.BoolNull(),
}

type InterfaceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Shutdown    types.Bool   `tfsdk:"shutdown"`
}

func SpanningTreeFromObjectValue(ctx context.Context, obj basetypes.ObjectValue) (SpanningTree, diag.Diagnostics) {
	var st SpanningTree
	diags := obj.As(ctx, &st, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		tflog.Error(ctx, "Failed to convert ObjectValue to SpanningTree")
		return SpanningTree{}, diags
	}
	return st, nil
}

func (st SpanningTree) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"portfast":   types.StringType,
		"bpdu_guard": types.BoolType,
	}
}

func (st SpanningTree) AttributeValues() map[string]attr.Value {
	return map[string]attr.Value{
		"portfast":   st.Portfast,
		"bpdu_guard": st.BpduGuard,
	}
}
