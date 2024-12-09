// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package store

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// AcceptFeedOwnership is the `acceptFeedOwnership` instruction.
type AcceptFeedOwnership struct {

	// [0] = [WRITE] feed
	//
	// [1] = [] proposedOwner
	//
	// [2] = [SIGNER] authority
	ag_solanago.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

// NewAcceptFeedOwnershipInstructionBuilder creates a new `AcceptFeedOwnership` instruction builder.
func NewAcceptFeedOwnershipInstructionBuilder() *AcceptFeedOwnership {
	nd := &AcceptFeedOwnership{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 3),
	}
	return nd
}

// SetFeedAccount sets the "feed" account.
func (inst *AcceptFeedOwnership) SetFeedAccount(feed ag_solanago.PublicKey) *AcceptFeedOwnership {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(feed).WRITE()
	return inst
}

// GetFeedAccount gets the "feed" account.
func (inst *AcceptFeedOwnership) GetFeedAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[0]
}

// SetProposedOwnerAccount sets the "proposedOwner" account.
func (inst *AcceptFeedOwnership) SetProposedOwnerAccount(proposedOwner ag_solanago.PublicKey) *AcceptFeedOwnership {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(proposedOwner)
	return inst
}

// GetProposedOwnerAccount gets the "proposedOwner" account.
func (inst *AcceptFeedOwnership) GetProposedOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[1]
}

// SetAuthorityAccount sets the "authority" account.
func (inst *AcceptFeedOwnership) SetAuthorityAccount(authority ag_solanago.PublicKey) *AcceptFeedOwnership {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(authority).SIGNER()
	return inst
}

// GetAuthorityAccount gets the "authority" account.
func (inst *AcceptFeedOwnership) GetAuthorityAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice[2]
}

func (inst AcceptFeedOwnership) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_AcceptFeedOwnership,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst AcceptFeedOwnership) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *AcceptFeedOwnership) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Feed is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.ProposedOwner is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Authority is not set")
		}
	}
	return nil
}

func (inst *AcceptFeedOwnership) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("AcceptFeedOwnership")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=3]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("         feed", inst.AccountMetaSlice[0]))
						accountsBranch.Child(ag_format.Meta("proposedOwner", inst.AccountMetaSlice[1]))
						accountsBranch.Child(ag_format.Meta("    authority", inst.AccountMetaSlice[2]))
					})
				})
		})
}

func (obj AcceptFeedOwnership) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *AcceptFeedOwnership) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewAcceptFeedOwnershipInstruction declares a new AcceptFeedOwnership instruction with the provided parameters and accounts.
func NewAcceptFeedOwnershipInstruction(
	// Accounts:
	feed ag_solanago.PublicKey,
	proposedOwner ag_solanago.PublicKey,
	authority ag_solanago.PublicKey) *AcceptFeedOwnership {
	return NewAcceptFeedOwnershipInstructionBuilder().
		SetFeedAccount(feed).
		SetProposedOwnerAccount(proposedOwner).
		SetAuthorityAccount(authority)
}
