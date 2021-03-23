package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	uuid "github.com/google/uuid"
)

// NewTxCmd is the top-level command for attribute CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"m"},
		Short:                      "Transaction commands for the metadata module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AddMetadataScopeCmd(),
		AddOsLocatorCmd(),
		RemoveOsLocatorCmd(),
		ModifyOsLocatorCmd(),
	)

	return txCmd
}

// AddMetadataScopeCmd creates a command for adding a metadata scope.
func AddMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-scope [scope-uuid] [spec-id] [owner-addresses] [data-access] [value-owner-address] [signers]",
		Short: "Add a metadata scope to the provenance blockchain",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scopeUUID, err := uuid.Parse(args[0])
			if err != nil {
				fmt.Printf("Invalid uuid for scope uuid: %s", args[0])
				return err
			}
			specUUID, err := uuid.Parse(args[1])
			if err != nil {
				fmt.Printf("Invalid uuid for specification uuid: %s", args[0])
				return err
			}

			specID := types.ScopeSpecMetadataAddress(specUUID)

			ownerAddresses := strings.Split(args[2], ",")
			owners := make([]types.Party, len(ownerAddresses))
			for i, ownerAddr := range ownerAddresses {
				owners[i] = types.Party{Address: ownerAddr, Role: types.PartyType_PARTY_TYPE_OWNER}
			}
			dataAccess := strings.Split(args[3], ",")
			valueOwnerAddress := args[4]

			signers := strings.Split(args[5], ",")
			for _, signer := range signers {
				_, err := sdk.AccAddressFromBech32(signer)
				if err != nil {
					fmt.Printf("signer address must be a Bech32 string: %v", err)
					return err
				}
			}

			scope := *types.NewScope(
				types.ScopeMetadataAddress(scopeUUID),
				specID,
				owners,
				dataAccess,
				valueOwnerAddress)

			if err := scope.ValidateBasic(); err != nil {
				fmt.Printf("Failed to validate scope %s : %v", scope.String(), err)
				return err
			}

			msg := types.NewMsgAddScopeRequest(
				scope,
				signers)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveMetadataScopeCmd creates a command for removing a scope.
func RemoveMetadataScopeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-scope [scope-address] [signers]",
		Short: "Remove a metadata scope to the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scopeUUID, err := uuid.Parse(args[0])
			if err != nil {
				fmt.Printf("Invalid uuid for scope id: %s", args[0])
				return err
			}

			scopeMetaAddress := types.ScopeMetadataAddress(scopeUUID)
			signers := strings.Split(args[1], ",")

			for _, signer := range signers {
				_, err := sdk.AccAddressFromBech32(signer)
				if err != nil {
					fmt.Printf("signer address must be a Bech32 string: %v", err)
					return err
				}
			}

			deleteScope := *types.NewMsgDeleteScopeRequest(scopeMetaAddress, signers)
			if err := deleteScope.ValidateBasic(); err != nil {
				fmt.Printf("Failed to validate remove scope %s : %v", deleteScope.String(), err)
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &deleteScope)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// RemoveMetadataScopeCmd creates a command for removing a scope.
func AddOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-locator [owner] [uri]",
		Short: "Add a uri to an owner address on the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, errAddr := sdk.AccAddressFromBech32(args[0]); errAddr != nil {
				fmt.Printf("failed to add locator for a given owner address, invalid address: %s\n", args[0])
				return fmt.Errorf("invalid address: %w", errAddr)
			}

			objectStoreLocator := types.ObjectStoreLocator{
				LocatorUri: args[1], Owner: args[0],
			}

			addOSLocator := *types.NewMsgBindOSLocatorRequest(objectStoreLocator)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &addOSLocator)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func RemoveOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-locator [owner] [uri]",
		Short: "Remove an os locator already associated owner address on the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, errAddr := sdk.AccAddressFromBech32(args[0]); errAddr != nil {
				fmt.Printf("failed to remove locator for a given owner address, invalid address: %s\n", args[0])
				return fmt.Errorf("invalid address: %w", errAddr)
			}

			objectStoreLocator := types.ObjectStoreLocator{
				LocatorUri: args[1], Owner: args[0],
			}

			deleteOSLocator := *types.NewMsgDeleteOSLocatorRequest(objectStoreLocator)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &deleteOSLocator)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func ModifyOsLocatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify-locator [owner] [uri]",
		Short: "Modify a uri already associated owner address on the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, errAddr := sdk.AccAddressFromBech32(args[0]); errAddr != nil {
				fmt.Printf("failed to add locator for a given owner address, invalid address: %s\n", args[0])
				return fmt.Errorf("invalid address: %w", errAddr)
			}
			if err != nil {
				fmt.Printf("Invalid uuid for scope id: %s", args[0])
				return err
			}

			objectStoreLocator := types.ObjectStoreLocator{
				LocatorUri: args[1], Owner: args[0],
			}

			modifyOSLocator := *types.NewMsgModifyOSLocatorRequest(objectStoreLocator)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &modifyOSLocator)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func AddRecordSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-record-specification [specification-id] [name] [input-specifications] [type-name] [result-types] [responsible-parties] [signers]",
		Short: "Add/Update metadata record specification to the provenance blockchain",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			specificationID, err := types.MetadataAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			if !specificationID.IsRecordSpecificationAddress() {
				return fmt.Errorf("invalid record specification id: %s", args[0])
			}

			inputs, err := inputSpecification(args[2])
			if err != nil {
				return err
			}

			resultType := definitionType(args[4])
			partyTypes := partyTypes(args[5])
			signers := strings.Split(args[6], ",")

			recordSpecification := types.RecordSpecification{
				SpecificationId:    specificationID,
				Name:               args[1],
				Inputs:             inputs,
				TypeName:           args[3],
				ResultType:         resultType,
				ResponsibleParties: partyTypes,
			}

			msg := *types.NewMsgAddRecordSpecificationRequest(recordSpecification, signers)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// inputSpecification converts cli delimited argument and converts it to InputSpecifications
func inputSpecification(cliDelimitedValue string) ([]*types.InputSpecification, error) {
	delimitedInputs := strings.Split(cliDelimitedValue, ";")
	inputs := make([]*types.InputSpecification, len(delimitedInputs))
	for i, delimitedInput := range delimitedInputs {
		values := strings.Split(delimitedInput, ",")
		if len(values) != 4 {
			return nil, fmt.Errorf("invalid number of values for input specification: %v", len(values))
		}
		name := values[0]
		typeName := values[1]
		switch s := strings.ToUpper(values[2]); s {
		case "RECORDID":
			recordID, err := types.MetadataAddressFromBech32(values[3])
			if err != nil {
				return nil, err
			}
			inputs[i] = &types.InputSpecification{
				Name: name,
				Source: &types.InputSpecification_RecordId{
					RecordId: recordID,
				},
				TypeName: typeName,
			}
		case "HASH":
			inputs[i] = &types.InputSpecification{
				Name:     name,
				Source:   &types.InputSpecification_Hash{Hash: values[3]},
				TypeName: typeName,
			}
		default:
			return nil, fmt.Errorf("incorrect source type for input specification: %s", s)
		}
	}
	return inputs, nil

}

func partyTypes(delimitedPartyTypes string) []types.PartyType {
	parties := strings.Split(delimitedPartyTypes, ",")
	partyTypes := make([]types.PartyType, len(parties))
	for i, party := range parties {
		partyValue := types.PartyType_value[fmt.Sprintf("PARTY_TYPE_%s", strings.ToUpper(party))]
		partyTypes[i] = types.PartyType(partyValue)
	}
	return partyTypes
}

func definitionType(cliValue string) types.DefinitionType {
	typeValue := types.DefinitionType_value[fmt.Sprintf("DEFINITION_TYPE_%s", strings.ToUpper(cliValue))]
	return types.DefinitionType(typeValue)
}

func RemoveRecordSpecificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-record-specification [owner] [uri]",
		Short: "Remove record specification from the provenance blockchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			//TODO: remove record specification
			msg := *types.NewMsgDeleteRecordSpecificationRequest()
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
