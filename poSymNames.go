//This file is autogenerated from https://github.com/immesys/bw2_pid/blob/master/allocations.yaml
package bw2bind


//Binary (0.0.0.0/4): Binary protocols
//This is a superclass for classes that are generally unreadable in their plain
//form and require translation.
const PONumBinary = 0
const PODFMaskBinary = `0.0.0.0/4`
const PODFBinary = `0.0.0.0`
const POMaskBinary = 4

//Text (64.0.0.0/4): Human readable text
//This is a superclass for classes that are moderately understandable if they
//are read directly in their binary form. Generally these are protocols that
//were designed specifically to be human readable.
const PONumText = 1073741824
const PODFMaskText = `64.0.0.0/4`
const PODFText = `64.0.0.0`
const POMaskText = 4

//BWRoutingObject (0.0.0.0/8): Bosswave Routing Object
//This class and schema block is reserved for bosswave routing objects
//represented using the full PID.
const PONumBWRoutingObject = 0
const PODFMaskBWRoutingObject = `0.0.0.0/8`
const PODFBWRoutingObject = `0.0.0.0`
const POMaskBWRoutingObject = 8

//Blob (1.0.0.0/8): Blob
//This is a class for schemas that do not use a public encoding format. In
//general it should be avoided. Schemas below this should include the key
//"readme" with a url to a description of the schema that is sufficiently
//detailed to allow for a developer to reverse engineer the protocol if
//required.
const PONumBlob = 16777216
const PODFMaskBlob = `1.0.0.0/8`
const PODFBlob = `1.0.0.0`
const POMaskBlob = 8

//MsgPack (2.0.0.0/8): MsgPack
//This class is for schemas that are represented in MsgPack
const PONumMsgPack = 33554432
const PODFMaskMsgPack = `2.0.0.0/8`
const PODFMsgPack = `2.0.0.0`
const POMaskMsgPack = 8

//CapnP (3.0.0.0/8): Captain Proto
//This class is for captain proto interfaces. Schemas below this should include
//the key "schema" with a url to their .capnp file
const PONumCapnP = 50331648
const PODFMaskCapnP = `3.0.0.0/8`
const PODFCapnP = `3.0.0.0`
const POMaskCapnP = 8

//JSON (65.0.0.0/8): JSON
//This class is for schemas that are represented in JSON
const PONumJSON = 1090519040
const PODFMaskJSON = `65.0.0.0/8`
const PODFJSON = `65.0.0.0`
const POMaskJSON = 8

//XML (66.0.0.0/8): XML
//This class is for schemas that are represented in XML
const PONumXML = 1107296256
const PODFMaskXML = `66.0.0.0/8`
const PODFXML = `66.0.0.0`
const POMaskXML = 8

//YAML (67.0.0.0/8): YAML
//This class is for schemas that are represented in YAML
const PONumYAML = 1124073472
const PODFMaskYAML = `67.0.0.0/8`
const PODFYAML = `67.0.0.0`
const POMaskYAML = 8

//LogDict (2.0.1.0/24): LogDict
//This class is for log messages encoded in msgpack
const PONumLogDict = 33554688
const PODFMaskLogDict = `2.0.1.0/24`
const PODFLogDict = `2.0.1.0`
const POMaskLogDict = 24

//HamiltonBase (2.0.4.0/24): Hamilton Messages
//This is the base class for messages used with the Hamilton motes. The only
//key guaranteed is "#" that contains a uint16 representation of the serial of
//the mote the message is destined for or originated from.
const PONumHamiltonBase = 33555456
const PODFMaskHamiltonBase = `2.0.4.0/24`
const PODFHamiltonBase = `2.0.4.0`
const POMaskHamiltonBase = 24

//HamiltonTelemetry (2.0.4.64/26): Hamilton Telemetry
//This object contains a "#" field for the serial number, as well as possibly
//containing an "A" field with a list of X, Y, and Z accelerometer values. A
//"T" field containing the temperature as an integer in degrees C multiplied by
//10000, and an "L" field containing the illumination in Lux.
const PONumHamiltonTelemetry = 33555520
const PODFMaskHamiltonTelemetry = `2.0.4.64/26`
const PODFHamiltonTelemetry = `2.0.4.64`
const POMaskHamiltonTelemetry = 26

//BinaryActuation (1.0.1.0/32): Binary actuation
//This payload object is one byte long, 0x00 for off, 0x01 for on.
const PONumBinaryActuation = 16777472
const PODFMaskBinaryActuation = `1.0.1.0/32`
const PODFBinaryActuation = `1.0.1.0`
const POMaskBinaryActuation = 32

//BWMessage (1.0.1.1/32): Packed Bosswave Message
//This object contains an entire signed and encoded bosswave message
const PONumBWMessage = 16777473
const PODFMaskBWMessage = `1.0.1.1/32`
const PODFBWMessage = `1.0.1.1`
const POMaskBWMessage = 32

//Double (1.0.2.0/32): Double
//This payload is an 8 byte long IEEE 754 double floating point value encoded
//in little endian. This should only be used if the semantic meaning is obvious
//in the context, otherwise a PID with a more specific semantic meaning should
//be used.
const PONumDouble = 16777728
const PODFMaskDouble = `1.0.2.0/32`
const PODFDouble = `1.0.2.0`
const POMaskDouble = 32

//SpawnpointLog (2.0.2.0/32): Spawnpoint stdout
//This contains stdout data from a spawnpoint container. It is a msgpacked
//dictionary that contains a "service" key, a "time" key (unix nano timestamp)
//and a "contents" key and a "spalias" key.
const PONumSpawnpointLog = 33554944
const PODFMaskSpawnpointLog = `2.0.2.0/32`
const PODFSpawnpointLog = `2.0.2.0`
const POMaskSpawnpointLog = 32

//SMetadata (2.0.3.1/32): Simple Metadata entry
//This contains a simple "val" string and "ts" int64 metadata entry. The key is
//determined by the URI. Other information MAY be present in the msgpacked
//object. The timestamp is used for merging metadata entries.
const PONumSMetadata = 33555201
const PODFMaskSMetadata = `2.0.3.1/32`
const PODFSMetadata = `2.0.3.1`
const POMaskSMetadata = 32

//String (64.0.1.0/32): String
//A plain string with no rigid semantic meaning. This can be thought of as a
//print statement. Anything that has semantic meaning like a process log should
//use a different schema.
const PONumString = 1073742080
const PODFMaskString = `64.0.1.0/32`
const PODFString = `64.0.1.0`
const POMaskString = 32

//FMDIntentString (64.0.1.1/32): FMD Intent String
//A plain string used as an intent for the follow-me display service.
const PONumFMDIntentString = 1073742081
const PODFMaskFMDIntentString = `64.0.1.1/32`
const PODFFMDIntentString = `64.0.1.1`
const POMaskFMDIntentString = 32

//SpawnpointConfig (67.0.2.0/32): SpawnPoint config
//A configuration file for SpawnPoint (github.com/immesys/spawnpoint)
const PONumSpawnpointConfig = 1124073984
const PODFMaskSpawnpointConfig = `67.0.2.0/32`
const PODFSpawnpointConfig = `67.0.2.0`
const POMaskSpawnpointConfig = 32

//SpawnpointHeartbeat (67.0.2.1/32): SpawnPoint heartbeat
//A heartbeat message from spawnpoint
const PONumSpawnpointHeartbeat = 1124073985
const PODFMaskSpawnpointHeartbeat = `67.0.2.1/32`
const PODFSpawnpointHeartbeat = `67.0.2.1`
const POMaskSpawnpointHeartbeat = 32

