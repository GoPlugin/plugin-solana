// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package store

import ag_binary "github.com/gagliardetto/binary"

type NewTransmission struct {
	Timestamp uint64
	Answer    ag_binary.Int128
}

func (obj NewTransmission) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Timestamp` param:
	err = encoder.Encode(obj.Timestamp)
	if err != nil {
		return err
	}
	// Serialize `Answer` param:
	err = encoder.Encode(obj.Answer)
	if err != nil {
		return err
	}
	return nil
}

func (obj *NewTransmission) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Timestamp`:
	err = decoder.Decode(&obj.Timestamp)
	if err != nil {
		return err
	}
	// Deserialize `Answer`:
	err = decoder.Decode(&obj.Answer)
	if err != nil {
		return err
	}
	return nil
}

type Round struct {
	RoundId   uint32
	Slot      uint64
	Timestamp uint32
	Answer    ag_binary.Int128
}

func (obj Round) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `RoundId` param:
	err = encoder.Encode(obj.RoundId)
	if err != nil {
		return err
	}
	// Serialize `Slot` param:
	err = encoder.Encode(obj.Slot)
	if err != nil {
		return err
	}
	// Serialize `Timestamp` param:
	err = encoder.Encode(obj.Timestamp)
	if err != nil {
		return err
	}
	// Serialize `Answer` param:
	err = encoder.Encode(obj.Answer)
	if err != nil {
		return err
	}
	return nil
}

func (obj *Round) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `RoundId`:
	err = decoder.Decode(&obj.RoundId)
	if err != nil {
		return err
	}
	// Deserialize `Slot`:
	err = decoder.Decode(&obj.Slot)
	if err != nil {
		return err
	}
	// Deserialize `Timestamp`:
	err = decoder.Decode(&obj.Timestamp)
	if err != nil {
		return err
	}
	// Deserialize `Answer`:
	err = decoder.Decode(&obj.Answer)
	if err != nil {
		return err
	}
	return nil
}

type Scope interface {
	isScope()
}

type scopeContainer struct {
	Enum            ag_binary.BorshEnum `borsh_enum:"true"`
	Version         Version
	Decimals        Decimals
	Description     Description
	RoundData       RoundData
	LatestRoundData LatestRoundData
	Aggregator      Aggregator
}

type Version uint8

func (obj Version) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *Version) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *Version) isScope() {}

type Decimals uint8

func (obj Decimals) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *Decimals) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *Decimals) isScope() {}

type Description uint8

func (obj Description) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *Description) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *Description) isScope() {}

type RoundData struct {
	RoundId uint32
}

func (obj RoundData) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `RoundId` param:
	err = encoder.Encode(obj.RoundId)
	if err != nil {
		return err
	}
	return nil
}

func (obj *RoundData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `RoundId`:
	err = decoder.Decode(&obj.RoundId)
	if err != nil {
		return err
	}
	return nil
}

func (_ *RoundData) isScope() {}

type LatestRoundData uint8

func (obj LatestRoundData) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *LatestRoundData) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *LatestRoundData) isScope() {}

type Aggregator uint8

func (obj Aggregator) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}

func (obj *Aggregator) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

func (_ *Aggregator) isScope() {}
