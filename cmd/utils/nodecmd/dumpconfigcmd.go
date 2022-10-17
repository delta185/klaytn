// Modifications Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from cmd/geth/config.go (2018/06/04).
// Modified and improved for the klaytn development.

package nodecmd

import (
	"bufio"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Shopify/sarama"
	"github.com/klaytn/klaytn/accounts"
	"github.com/klaytn/klaytn/accounts/keystore"
	"github.com/klaytn/klaytn/api/debug"
	"github.com/klaytn/klaytn/blockchain"
	"github.com/klaytn/klaytn/cmd/utils"
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/common/fdlimit"
	"github.com/klaytn/klaytn/crypto"
	"github.com/klaytn/klaytn/datasync/chaindatafetcher"
	"github.com/klaytn/klaytn/datasync/chaindatafetcher/kafka"
	"github.com/klaytn/klaytn/datasync/chaindatafetcher/kas"
	"github.com/klaytn/klaytn/datasync/dbsyncer"
	"github.com/klaytn/klaytn/datasync/downloader"
	"github.com/klaytn/klaytn/log"
	"github.com/klaytn/klaytn/networks/p2p"
	"github.com/klaytn/klaytn/networks/p2p/discover"
	"github.com/klaytn/klaytn/networks/p2p/nat"
	"github.com/klaytn/klaytn/networks/p2p/netutil"
	"github.com/klaytn/klaytn/networks/rpc"
	"github.com/klaytn/klaytn/node"
	"github.com/klaytn/klaytn/node/cn"
	"github.com/klaytn/klaytn/node/cn/filters"
	"github.com/klaytn/klaytn/node/sc"
	"github.com/klaytn/klaytn/params"
	"github.com/klaytn/klaytn/storage/database"
	"github.com/klaytn/klaytn/storage/statedb"
	"github.com/naoina/toml"
	"gopkg.in/urfave/cli.v1"
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type klayConfig struct {
	CN               cn.Config
	Node             node.Config
	DB               dbsyncer.DBConfig
	ChainDataFetcher chaindatafetcher.ChainDataFetcherConfig
	ServiceChain     sc.SCConfig
}

// GetDumpConfigCommand returns cli.Command `dumpconfig` whose flags are initialized with nodeFlags and rpcFlags.
func GetDumpConfigCommand(nodeFlags, rpcFlags []cli.Flag) cli.Command {
	return cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(nodeFlags, rpcFlags...)),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}
}

func loadConfig(file string, cfg *klayConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "klay", "shh", "eth")
	cfg.WSModules = append(cfg.WSModules, "klay", "shh", "eth")
	cfg.IPCPath = "klay.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, klayConfig) {
	// Load defaults.
	cfg := klayConfig{
		CN:               *cn.GetDefaultConfig(),
		Node:             defaultNodeConfig(),
		DB:               *dbsyncer.DefaultDBConfig(),
		ChainDataFetcher: *chaindatafetcher.DefaultChainDataFetcherConfig(),
		ServiceChain:     *sc.DefaultServiceChainConfig(),
	}

	// Load config file.
	if file := ctx.GlobalString(utils.ConfigFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			log.Fatalf("%v", err)
		}
	}

	// Apply flags.
	cfg.SetNodeConfig(ctx)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		log.Fatalf("Failed to create the protocol stack: %v", err)
	}
	cfg.SetKlayConfig(ctx, stack)

	cfg.setDBSyncerConfig(ctx)
	cfg.setChainDataFetcherConfig(ctx)
	cfg.setServiceChainConfig(ctx)

	// utils.SetShhConfig(ctx, stack, &cfg.Shh)
	// utils.SetDashboardConfig(ctx, &cfg.Dashboard)

	return stack, cfg
}

func SetP2PConfig(ctx *cli.Context, cfg *p2p.Config) {
	setNodeKey(ctx, cfg)
	setNAT(ctx, cfg)
	setListenAddress(ctx, cfg)

	var nodeType string
	if ctx.GlobalIsSet(utils.NodeTypeFlag.Name) {
		nodeType = ctx.GlobalString(utils.NodeTypeFlag.Name)
	} else {
		nodeType = utils.NodeTypeFlag.Value
	}

	cfg.ConnectionType = convertNodeType(nodeType)
	if cfg.ConnectionType == common.UNKNOWNNODE {
		logger.Crit("Unknown node type", "nodetype", nodeType)
	}
	logger.Info("Setting connection type", "nodetype", nodeType, "conntype", cfg.ConnectionType)

	// set bootnodes via this function by check specified parameters
	setBootstrapNodes(ctx, cfg)

	if ctx.GlobalIsSet(utils.MaxConnectionsFlag.Name) {
		cfg.MaxPhysicalConnections = ctx.GlobalInt(utils.MaxConnectionsFlag.Name)
	}
	logger.Info("Setting MaxPhysicalConnections", "MaxPhysicalConnections", cfg.MaxPhysicalConnections)

	if ctx.GlobalIsSet(utils.MaxPendingPeersFlag.Name) {
		cfg.MaxPendingPeers = ctx.GlobalInt(utils.MaxPendingPeersFlag.Name)
	}

	cfg.NoDiscovery = ctx.GlobalIsSet(utils.NoDiscoverFlag.Name)

	cfg.RWTimerConfig = p2p.RWTimerConfig{}
	cfg.RWTimerConfig.Interval = ctx.GlobalUint64(utils.RWTimerIntervalFlag.Name)
	cfg.RWTimerConfig.WaitTime = ctx.GlobalDuration(utils.RWTimerWaitTimeFlag.Name)

	if netrestrict := ctx.GlobalString(utils.NetrestrictFlag.Name); netrestrict != "" {
		list, err := netutil.ParseNetlist(netrestrict)
		if err != nil {
			log.Fatalf("Option %q: %v", utils.NetrestrictFlag.Name, err)
		}
		cfg.NetRestrict = list
	}

	common.MaxRequestContentLength = ctx.GlobalInt(utils.MaxRequestContentLengthFlag.Name)

	cfg.NetworkID, _ = getNetworkId(ctx)
}

// setNodeKey creates a node key from set command line flags, either loading it
// from a file or as a specified hex value. If neither flags were provided, this
// method returns nil and an emphemeral key is to be generated.
func setNodeKey(ctx *cli.Context, cfg *p2p.Config) {
	var (
		hex  = ctx.GlobalString(utils.NodeKeyHexFlag.Name)
		file = ctx.GlobalString(utils.NodeKeyFileFlag.Name)
		key  *ecdsa.PrivateKey
		err  error
	)
	switch {
	case file != "" && hex != "":
		log.Fatalf("Options %q and %q are mutually exclusive", utils.NodeKeyFileFlag.Name, utils.NodeKeyHexFlag.Name)
	case file != "":
		if key, err = crypto.LoadECDSA(file); err != nil {
			log.Fatalf("Option %q: %v", utils.NodeKeyFileFlag.Name, err)
		}
		cfg.PrivateKey = key
	case hex != "":
		if key, err = crypto.HexToECDSA(hex); err != nil {
			log.Fatalf("Option %q: %v", utils.NodeKeyHexFlag.Name, err)
		}
		cfg.PrivateKey = key
	}
}

// setNAT creates a port mapper from command line flags.
func setNAT(ctx *cli.Context, cfg *p2p.Config) {
	if ctx.GlobalIsSet(utils.NATFlag.Name) {
		natif, err := nat.Parse(ctx.GlobalString(utils.NATFlag.Name))
		if err != nil {
			log.Fatalf("Option %s: %v", utils.NATFlag.Name, err)
		}
		cfg.NAT = natif
	}
}

// setListenAddress creates a TCP listening address string from set command
// line flags.
func setListenAddress(ctx *cli.Context, cfg *p2p.Config) {
	if ctx.GlobalIsSet(utils.ListenPortFlag.Name) {
		cfg.ListenAddr = fmt.Sprintf(":%d", ctx.GlobalInt(utils.ListenPortFlag.Name))
	}

	if ctx.GlobalBool(utils.MultiChannelUseFlag.Name) {
		cfg.EnableMultiChannelServer = true
		SubListenAddr := fmt.Sprintf(":%d", ctx.GlobalInt(utils.SubListenPortFlag.Name))
		cfg.SubListenAddr = []string{SubListenAddr}
	}
}

func convertNodeType(nodetype string) common.ConnType {
	switch strings.ToLower(nodetype) {
	case "cn", "scn":
		return common.CONSENSUSNODE
	case "pn", "spn":
		return common.PROXYNODE
	case "en", "sen":
		return common.ENDPOINTNODE
	default:
		return common.UNKNOWNNODE
	}
}

// setBootstrapNodes creates a list of bootstrap nodes from the command line
// flags, reverting to pre-configured ones if none have been specified.
func setBootstrapNodes(ctx *cli.Context, cfg *p2p.Config) {
	var urls []string
	switch {
	case ctx.GlobalIsSet(utils.BootnodesFlag.Name):
		logger.Info("Customized bootnodes are set")
		urls = strings.Split(ctx.GlobalString(utils.BootnodesFlag.Name), ",")
	case ctx.GlobalIsSet(utils.CypressFlag.Name):
		logger.Info("Cypress bootnodes are set")
		urls = params.MainnetBootnodes[cfg.ConnectionType].Addrs
	case ctx.GlobalIsSet(utils.BaobabFlag.Name):
		logger.Info("Baobab bootnodes are set")
		// set pre-configured bootnodes when 'baobab' option was enabled
		urls = params.BaobabBootnodes[cfg.ConnectionType].Addrs
	case cfg.BootstrapNodes != nil:
		return // already set, don't apply defaults.
	case !ctx.GlobalIsSet(utils.NetworkIdFlag.Name):
		if utils.NodeTypeFlag.Value != "scn" && utils.NodeTypeFlag.Value != "spn" && utils.NodeTypeFlag.Value != "sen" {
			logger.Info("Cypress bootnodes are set")
			urls = params.MainnetBootnodes[cfg.ConnectionType].Addrs
		}
	}

	cfg.BootstrapNodes = make([]*discover.Node, 0, len(urls))
	for _, url := range urls {
		node, err := discover.ParseNode(url)
		if err != nil {
			logger.Error("Bootstrap URL invalid", "kni", url, "err", err)
			continue
		}
		if node.NType == discover.NodeTypeUnknown {
			logger.Debug("setBootstrapNode: set nodetype as bn from unknown", "nodeid", node.ID)
			node.NType = discover.NodeTypeBN
		}
		logger.Info("Bootnode - Add Seed", "Node", node)
		cfg.BootstrapNodes = append(cfg.BootstrapNodes, node)
	}
}

// SetNodeConfig applies node-related command line flags to the config.
func (kCfg *klayConfig) SetNodeConfig(ctx *cli.Context) {
	cfg := &kCfg.Node
	SetP2PConfig(ctx, &cfg.P2P)
	setIPC(ctx, cfg)

	// httptype is http or fasthttp
	if ctx.GlobalIsSet(utils.SrvTypeFlag.Name) {
		cfg.HTTPServerType = ctx.GlobalString(utils.SrvTypeFlag.Name)
	}

	setHTTP(ctx, cfg)
	setWS(ctx, cfg)
	setgRPC(ctx, cfg)
	setAPIConfig(ctx)
	setNodeUserIdent(ctx, cfg)

	if dbtype := database.DBType(ctx.GlobalString(utils.DbTypeFlag.Name)).ToValid(); len(dbtype) != 0 {
		cfg.DBType = dbtype
	} else {
		logger.Crit("invalid dbtype", "dbtype", ctx.GlobalString(utils.DbTypeFlag.Name))
	}
	cfg.DataDir = ctx.GlobalString(utils.DataDirFlag.Name)

	if ctx.GlobalIsSet(utils.KeyStoreDirFlag.Name) {
		cfg.KeyStoreDir = ctx.GlobalString(utils.KeyStoreDirFlag.Name)
	}
	if ctx.GlobalIsSet(utils.LightKDFFlag.Name) {
		cfg.UseLightweightKDF = ctx.GlobalBool(utils.LightKDFFlag.Name)
	}
	if ctx.GlobalIsSet(utils.RPCNonEthCompatibleFlag.Name) {
		rpc.NonEthCompatible = ctx.GlobalBool(utils.RPCNonEthCompatibleFlag.Name)
	}
}

// setHTTP creates the HTTP RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setHTTP(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(utils.RPCEnabledFlag.Name) && cfg.HTTPHost == "" {
		cfg.HTTPHost = "127.0.0.1"
		if ctx.GlobalIsSet(utils.RPCListenAddrFlag.Name) {
			cfg.HTTPHost = ctx.GlobalString(utils.RPCListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(utils.RPCPortFlag.Name) {
		cfg.HTTPPort = ctx.GlobalInt(utils.RPCPortFlag.Name)
	}
	if ctx.GlobalIsSet(utils.RPCCORSDomainFlag.Name) {
		cfg.HTTPCors = utils.SplitAndTrim(ctx.GlobalString(utils.RPCCORSDomainFlag.Name))
	}
	if ctx.GlobalIsSet(utils.RPCApiFlag.Name) {
		cfg.HTTPModules = utils.SplitAndTrim(ctx.GlobalString(utils.RPCApiFlag.Name))
	}
	if ctx.GlobalIsSet(utils.RPCVirtualHostsFlag.Name) {
		cfg.HTTPVirtualHosts = utils.SplitAndTrim(ctx.GlobalString(utils.RPCVirtualHostsFlag.Name))
	}
	if ctx.GlobalIsSet(utils.RPCConcurrencyLimit.Name) {
		rpc.ConcurrencyLimit = ctx.GlobalInt(utils.RPCConcurrencyLimit.Name)
		logger.Info("Set the concurrency limit of RPC-HTTP server", "limit", rpc.ConcurrencyLimit)
	}
	if ctx.GlobalIsSet(utils.RPCReadTimeout.Name) {
		cfg.HTTPTimeouts.ReadTimeout = time.Duration(ctx.GlobalInt(utils.RPCReadTimeout.Name)) * time.Second
	}
	if ctx.GlobalIsSet(utils.RPCWriteTimeoutFlag.Name) {
		cfg.HTTPTimeouts.WriteTimeout = time.Duration(ctx.GlobalInt(utils.RPCWriteTimeoutFlag.Name)) * time.Second
	}
	if ctx.GlobalIsSet(utils.RPCIdleTimeoutFlag.Name) {
		cfg.HTTPTimeouts.IdleTimeout = time.Duration(ctx.GlobalInt(utils.RPCIdleTimeoutFlag.Name)) * time.Second
	}
	if ctx.GlobalIsSet(utils.RPCExecutionTimeoutFlag.Name) {
		cfg.HTTPTimeouts.ExecutionTimeout = time.Duration(ctx.GlobalInt(utils.RPCExecutionTimeoutFlag.Name)) * time.Second
	}
}

// setWS creates the WebSocket RPC listener interface string from the set
// command line flags, returning empty if the HTTP endpoint is disabled.
func setWS(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(utils.WSEnabledFlag.Name) && cfg.WSHost == "" {
		cfg.WSHost = "127.0.0.1"
		if ctx.GlobalIsSet(utils.WSListenAddrFlag.Name) {
			cfg.WSHost = ctx.GlobalString(utils.WSListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(utils.WSPortFlag.Name) {
		cfg.WSPort = ctx.GlobalInt(utils.WSPortFlag.Name)
	}
	if ctx.GlobalIsSet(utils.WSAllowedOriginsFlag.Name) {
		cfg.WSOrigins = utils.SplitAndTrim(ctx.GlobalString(utils.WSAllowedOriginsFlag.Name))
	}
	if ctx.GlobalIsSet(utils.WSApiFlag.Name) {
		cfg.WSModules = utils.SplitAndTrim(ctx.GlobalString(utils.WSApiFlag.Name))
	}
	rpc.MaxSubscriptionPerWSConn = int32(ctx.GlobalInt(utils.WSMaxSubscriptionPerConn.Name))
	rpc.WebsocketReadDeadline = ctx.GlobalInt64(utils.WSReadDeadLine.Name)
	rpc.WebsocketWriteDeadline = ctx.GlobalInt64(utils.WSWriteDeadLine.Name)
	rpc.MaxWebsocketConnections = int32(ctx.GlobalInt(utils.WSMaxConnections.Name))
}

// setIPC creates an IPC path configuration from the set command line flags,
// returning an empty string if IPC was explicitly disabled, or the set path.
func setIPC(ctx *cli.Context, cfg *node.Config) {
	CheckExclusive(ctx, utils.IPCDisabledFlag, utils.IPCPathFlag)
	switch {
	case ctx.GlobalBool(utils.IPCDisabledFlag.Name):
		cfg.IPCPath = ""
	case ctx.GlobalIsSet(utils.IPCPathFlag.Name):
		cfg.IPCPath = ctx.GlobalString(utils.IPCPathFlag.Name)
	}
}

// setgRPC creates the gRPC listener interface string from the set
// command line flags, returning empty if the gRPC endpoint is disabled.
func setgRPC(ctx *cli.Context, cfg *node.Config) {
	if ctx.GlobalBool(utils.GRPCEnabledFlag.Name) && cfg.GRPCHost == "" {
		cfg.GRPCHost = "127.0.0.1"
		if ctx.GlobalIsSet(utils.GRPCListenAddrFlag.Name) {
			cfg.GRPCHost = ctx.GlobalString(utils.GRPCListenAddrFlag.Name)
		}
	}

	if ctx.GlobalIsSet(utils.GRPCPortFlag.Name) {
		cfg.GRPCPort = ctx.GlobalInt(utils.GRPCPortFlag.Name)
	}
}

// setAPIConfig sets configurations for specific APIs.
func setAPIConfig(ctx *cli.Context) {
	filters.GetLogsDeadline = ctx.GlobalDuration(utils.APIFilterGetLogsDeadlineFlag.Name)
	filters.GetLogsMaxItems = ctx.GlobalInt(utils.APIFilterGetLogsMaxItemsFlag.Name)
}

// setNodeUserIdent creates the user identifier from CLI flags.
func setNodeUserIdent(ctx *cli.Context, cfg *node.Config) {
	if identity := ctx.GlobalString(utils.IdentityFlag.Name); len(identity) > 0 {
		cfg.UserIdent = identity
	}
}

// CheckExclusive verifies that only a single instance of the provided flags was
// set by the user. Each flag might optionally be followed by a string type to
// specialize it further.
func CheckExclusive(ctx *cli.Context, args ...interface{}) {
	set := make([]string, 0, 1)
	for i := 0; i < len(args); i++ {
		// Make sure the next argument is a flag and skip if not set
		flag, ok := args[i].(cli.Flag)
		if !ok {
			panic(fmt.Sprintf("invalid argument, not cli.Flag type: %T", args[i]))
		}
		// Check if next arg extends current and expand its name if so
		name := flag.GetName()

		if i+1 < len(args) {
			switch option := args[i+1].(type) {
			case string:
				// Extended flag, expand the name and shift the arguments
				if ctx.GlobalString(flag.GetName()) == option {
					name += "=" + option
				}
				i++

			case cli.Flag:
			default:
				panic(fmt.Sprintf("invalid argument, not cli.Flag or string extension: %T", args[i+1]))
			}
		}
		// Mark the flag if it's set
		if ctx.GlobalIsSet(flag.GetName()) {
			set = append(set, "--"+name)
		}
	}
	if len(set) > 1 {
		log.Fatalf("Flags %v can't be used at the same time", strings.Join(set, ", "))
	}
}

// SetKlayConfig applies klay-related command line flags to the config.
func (kCfg *klayConfig) SetKlayConfig(ctx *cli.Context, stack *node.Node) {
	// TODO-Klaytn-Bootnode: better have to check conflicts about network flags when we add Klaytn's `mainnet` parameter
	// checkExclusive(ctx, DeveloperFlag, TestnetFlag, RinkebyFlag)
	cfg := &kCfg.CN
	raiseFDLimit()

	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	setServiceChainSigner(ctx, ks, cfg)
	setRewardbase(ctx, ks, cfg)
	setTxPool(ctx, &cfg.TxPool)

	if ctx.GlobalIsSet(utils.SyncModeFlag.Name) {
		cfg.SyncMode = *utils.GlobalTextMarshaler(ctx, utils.SyncModeFlag.Name).(*downloader.SyncMode)
		if cfg.SyncMode != downloader.FullSync && cfg.SyncMode != downloader.SnapSync {
			log.Fatalf("Full Sync or Snap Sync (prototype) is supported only!")
		}
		if cfg.SyncMode == downloader.SnapSync {
			logger.Info("Snap sync requested, enabling --snapshot")
			ctx.Set(utils.SnapshotFlag.Name, "true")
		} else {
			cfg.SnapshotCacheSize = 0 // Disabled
		}
	}

	if ctx.GlobalBool(utils.KESNodeTypeServiceFlag.Name) {
		cfg.FetcherDisable = true
		cfg.DownloaderDisable = true
		cfg.WorkerDisable = true
	}

	if utils.NetworkTypeFlag.Value == SCNNetworkType && kCfg.ServiceChain.EnabledSubBridge {
		cfg.NoAccountCreation = !ctx.GlobalBool(utils.ServiceChainNewAccountFlag.Name)
	}

	cfg.NetworkId, cfg.IsPrivate = getNetworkId(ctx)

	if dbtype := database.DBType(ctx.GlobalString(utils.DbTypeFlag.Name)).ToValid(); len(dbtype) != 0 {
		cfg.DBType = dbtype
	} else {
		logger.Crit("invalid dbtype", "dbtype", ctx.GlobalString(utils.DbTypeFlag.Name))
	}
	cfg.SingleDB = ctx.GlobalIsSet(utils.SingleDBFlag.Name)
	cfg.NumStateTrieShards = ctx.GlobalUint(utils.NumStateTrieShardsFlag.Name)
	if !database.IsPow2(cfg.NumStateTrieShards) {
		log.Fatalf("%v should be power of 2 but %v is not!", utils.NumStateTrieShardsFlag.Name, cfg.NumStateTrieShards)
	}

	cfg.OverwriteGenesis = ctx.GlobalBool(utils.OverwriteGenesisFlag.Name)
	cfg.StartBlockNumber = ctx.GlobalUint64(utils.StartBlockNumberFlag.Name)

	cfg.LevelDBCompression = database.LevelDBCompressionType(ctx.GlobalInt(utils.LevelDBCompressionTypeFlag.Name))
	cfg.LevelDBBufferPool = !ctx.GlobalIsSet(utils.LevelDBNoBufferPoolFlag.Name)
	cfg.EnableDBPerfMetrics = !ctx.GlobalIsSet(utils.DBNoPerformanceMetricsFlag.Name)
	cfg.LevelDBCacheSize = ctx.GlobalInt(utils.LevelDBCacheSizeFlag.Name)

	cfg.DynamoDBConfig.TableName = ctx.GlobalString(utils.DynamoDBTableNameFlag.Name)
	cfg.DynamoDBConfig.Region = ctx.GlobalString(utils.DynamoDBRegionFlag.Name)
	cfg.DynamoDBConfig.IsProvisioned = ctx.GlobalBool(utils.DynamoDBIsProvisionedFlag.Name)
	cfg.DynamoDBConfig.ReadCapacityUnits = ctx.GlobalInt64(utils.DynamoDBReadCapacityFlag.Name)
	cfg.DynamoDBConfig.WriteCapacityUnits = ctx.GlobalInt64(utils.DynamoDBWriteCapacityFlag.Name)
	cfg.DynamoDBConfig.ReadOnly = ctx.GlobalBool(utils.DynamoDBReadOnlyFlag.Name)

	if gcmode := ctx.GlobalString(utils.GCModeFlag.Name); gcmode != "full" && gcmode != "archive" {
		log.Fatalf("--%s must be either 'full' or 'archive'", utils.GCModeFlag.Name)
	}
	cfg.NoPruning = ctx.GlobalString(utils.GCModeFlag.Name) == "archive"
	logger.Info("Archiving mode of this node", "isArchiveMode", cfg.NoPruning)

	cfg.AnchoringPeriod = ctx.GlobalUint64(utils.AnchoringPeriodFlag.Name)
	cfg.SentChainTxsLimit = ctx.GlobalUint64(utils.SentChainTxsLimit.Name)

	cfg.TrieCacheSize = ctx.GlobalInt(utils.TrieMemoryCacheSizeFlag.Name)
	common.DefaultCacheType = common.CacheType(ctx.GlobalInt(utils.CacheTypeFlag.Name))
	cfg.TrieBlockInterval = ctx.GlobalUint(utils.TrieBlockIntervalFlag.Name)
	cfg.TriesInMemory = ctx.GlobalUint64(utils.TriesInMemoryFlag.Name)

	if ctx.GlobalIsSet(utils.CacheScaleFlag.Name) {
		common.CacheScale = ctx.GlobalInt(utils.CacheScaleFlag.Name)
	}
	if ctx.GlobalIsSet(utils.CacheUsageLevelFlag.Name) {
		cacheUsageLevelFlag := ctx.GlobalString(utils.CacheUsageLevelFlag.Name)
		if scaleByCacheUsageLevel, err := common.GetScaleByCacheUsageLevel(cacheUsageLevelFlag); err != nil {
			logger.Crit("Incorrect CacheUsageLevelFlag value", "error", err, "CacheUsageLevelFlag", cacheUsageLevelFlag)
		} else {
			common.ScaleByCacheUsageLevel = scaleByCacheUsageLevel
		}
	}
	if ctx.GlobalIsSet(utils.MemorySizeFlag.Name) {
		physicalMemory := common.TotalPhysicalMemGB
		common.TotalPhysicalMemGB = ctx.GlobalInt(utils.MemorySizeFlag.Name)
		logger.Info("Physical memory has been replaced by user settings", "PhysicalMemory(GB)", physicalMemory, "UserSetting(GB)", common.TotalPhysicalMemGB)
	} else {
		logger.Debug("Memory settings", "PhysicalMemory(GB)", common.TotalPhysicalMemGB)
	}

	if ctx.GlobalIsSet(utils.DocRootFlag.Name) {
		cfg.DocRoot = ctx.GlobalString(utils.DocRootFlag.Name)
	}
	if ctx.GlobalIsSet(utils.ExtraDataFlag.Name) {
		cfg.ExtraData = []byte(ctx.GlobalString(utils.ExtraDataFlag.Name))
	}

	cfg.SenderTxHashIndexing = ctx.GlobalIsSet(utils.SenderTxHashIndexingFlag.Name)
	cfg.ParallelDBWrite = !ctx.GlobalIsSet(utils.NoParallelDBWriteFlag.Name)
	cfg.TrieNodeCacheConfig = statedb.TrieNodeCacheConfig{
		CacheType: statedb.TrieNodeCacheType(ctx.GlobalString(utils.TrieNodeCacheTypeFlag.
			Name)).ToValid(),
		NumFetcherPrefetchWorker:  ctx.GlobalInt(utils.NumFetcherPrefetchWorkerFlag.Name),
		UseSnapshotForPrefetch:    ctx.GlobalBool(utils.UseSnapshotForPrefetchFlag.Name),
		LocalCacheSizeMiB:         ctx.GlobalInt(utils.TrieNodeCacheLimitFlag.Name),
		FastCacheFileDir:          ctx.GlobalString(utils.DataDirFlag.Name) + "/fastcache",
		FastCacheSavePeriod:       ctx.GlobalDuration(utils.TrieNodeCacheSavePeriodFlag.Name),
		RedisEndpoints:            ctx.GlobalStringSlice(utils.TrieNodeCacheRedisEndpointsFlag.Name),
		RedisClusterEnable:        ctx.GlobalBool(utils.TrieNodeCacheRedisClusterFlag.Name),
		RedisPublishBlockEnable:   ctx.GlobalBool(utils.TrieNodeCacheRedisPublishBlockFlag.Name),
		RedisSubscribeBlockEnable: ctx.GlobalBool(utils.TrieNodeCacheRedisSubscribeBlockFlag.Name),
	}

	if ctx.GlobalIsSet(utils.VMEnableDebugFlag.Name) {
		// TODO(fjl): force-enable this in --dev mode
		cfg.EnablePreimageRecording = ctx.GlobalBool(utils.VMEnableDebugFlag.Name)
	}
	if ctx.GlobalIsSet(utils.VMLogTargetFlag.Name) {
		if _, err := debug.Handler.SetVMLogTarget(ctx.GlobalInt(utils.VMLogTargetFlag.Name)); err != nil {
			logger.Warn("Incorrect vmlog value", "err", err)
		}
	}
	cfg.EnableInternalTxTracing = ctx.GlobalIsSet(utils.VMTraceInternalTxFlag.Name)

	cfg.AutoRestartFlag = ctx.GlobalBool(utils.AutoRestartFlag.Name)
	cfg.RestartTimeOutFlag = ctx.GlobalDuration(utils.RestartTimeOutFlag.Name)
	cfg.DaemonPathFlag = ctx.GlobalString(utils.DaemonPathFlag.Name)

	if ctx.GlobalIsSet(utils.RPCGlobalGasCap.Name) {
		cfg.RPCGasCap = new(big.Int).SetUint64(ctx.GlobalUint64(utils.RPCGlobalGasCap.Name))
	}

	if ctx.GlobalIsSet(utils.RPCGlobalEthTxFeeCapFlag.Name) {
		cfg.RPCTxFeeCap = ctx.GlobalFloat64(utils.RPCGlobalEthTxFeeCapFlag.Name)
	}

	// Only CNs could set BlockGenerationIntervalFlag and BlockGenerationTimeLimitFlag
	if ctx.GlobalIsSet(utils.BlockGenerationIntervalFlag.Name) {
		params.BlockGenerationInterval = ctx.GlobalInt64(utils.BlockGenerationIntervalFlag.Name)
		if params.BlockGenerationInterval < 1 {
			logger.Crit("Block generation interval should be equal or larger than 1", "interval", params.BlockGenerationInterval)
		}
	}
	if ctx.GlobalIsSet(utils.BlockGenerationTimeLimitFlag.Name) {
		params.BlockGenerationTimeLimit = ctx.GlobalDuration(utils.BlockGenerationTimeLimitFlag.Name)
	}

	params.OpcodeComputationCostLimit = ctx.GlobalUint64(utils.OpcodeComputationCostLimitFlag.Name)

	if ctx.GlobalIsSet(utils.SnapshotFlag.Name) {
		cfg.SnapshotCacheSize = ctx.GlobalInt(utils.SnapshotCacheSizeFlag.Name)
		if cfg.StartBlockNumber != 0 {
			logger.Crit("State snapshot should not be used with --start-block-num", "num", cfg.StartBlockNumber)
		}
		logger.Info("State snapshot is enabled", "cache-size (MB)", cfg.SnapshotCacheSize)
	} else {
		cfg.SnapshotCacheSize = 0 // snapshot disabled
	}

	// Override any default configs for hard coded network.
	// TODO-Klaytn-Bootnode: Discuss and add `baobab` test network's genesis block
	/*
		if ctx.GlobalBool(TestnetFlag.Name) {
			if !ctx.GlobalIsSet(NetworkIdFlag.Name) {
				cfg.NetworkId = 3
			}
			cfg.Genesis = blockchain.DefaultBaobabGenesisBlock()
		}
	*/
	// Set the Tx resending related configuration variables
	setTxResendConfig(ctx, cfg)
}

// raiseFDLimit increases the file descriptor limit to process's maximum value
func raiseFDLimit() {
	limit, err := fdlimit.Maximum()
	if err != nil {
		logger.Error("Failed to read maximum fd. you may suffer fd exhaustion", "err", err)
		return
	}
	raised, err := fdlimit.Raise(uint64(limit))
	if err != nil {
		logger.Warn("Failed to increase fd limit. you may suffer fd exhaustion", "err", err)
		return
	}
	logger.Info("Raised fd limit to process's maximum value", "fd", raised)
}

// setServiceChainSigner retrieves the service chain signer either from the directly specified
// command line flags or from the keystore if CLI indexed.
func setServiceChainSigner(ctx *cli.Context, ks *keystore.KeyStore, cfg *cn.Config) {
	if ctx.GlobalIsSet(utils.ServiceChainSignerFlag.Name) {
		account, err := makeAddress(ks, ctx.GlobalString(utils.ServiceChainSignerFlag.Name))
		if err != nil {
			log.Fatalf("Option %q: %v", utils.ServiceChainSignerFlag.Name, err)
		}
		cfg.ServiceChainSigner = account.Address
	}
}

// setRewardbase retrieves the rewardbase either from the directly specified
// command line flags or from the keystore if CLI indexed.
func setRewardbase(ctx *cli.Context, ks *keystore.KeyStore, cfg *cn.Config) {
	if ctx.GlobalIsSet(utils.RewardbaseFlag.Name) {
		account, err := makeAddress(ks, ctx.GlobalString(utils.RewardbaseFlag.Name))
		if err != nil {
			log.Fatalf("Option %q: %v", utils.RewardbaseFlag.Name, err)
		}
		cfg.Rewardbase = account.Address
	}
}

// MakeAddress converts an account specified directly as a hex encoded string or
// a key index in the key store to an internal account representation.
func makeAddress(ks *keystore.KeyStore, account string) (accounts.Account, error) {
	// If the specified account is a valid address, return it
	if common.IsHexAddress(account) {
		return accounts.Account{Address: common.HexToAddress(account)}, nil
	}
	// Otherwise try to interpret the account as a keystore index
	index, err := strconv.Atoi(account)
	if err != nil || index < 0 {
		return accounts.Account{}, fmt.Errorf("invalid account address or index %q", account)
	}
	logger.Warn("Use explicit addresses! Referring to accounts by order in the keystore folder is dangerous and will be deprecated!")

	accs := ks.Accounts()
	if len(accs) <= index {
		return accounts.Account{}, fmt.Errorf("index %d higher than number of accounts %d", index, len(accs))
	}
	return accs[index], nil
}

func setTxPool(ctx *cli.Context, cfg *blockchain.TxPoolConfig) {
	if ctx.GlobalIsSet(utils.TxPoolNoLocalsFlag.Name) {
		cfg.NoLocals = ctx.GlobalBool(utils.TxPoolNoLocalsFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolAllowLocalAnchorTxFlag.Name) {
		cfg.AllowLocalAnchorTx = ctx.GlobalBool(utils.TxPoolAllowLocalAnchorTxFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolDenyRemoteTxFlag.Name) {
		cfg.DenyRemoteTx = ctx.GlobalBool(utils.TxPoolDenyRemoteTxFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolJournalFlag.Name) {
		cfg.Journal = ctx.GlobalString(utils.TxPoolJournalFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolJournalIntervalFlag.Name) {
		cfg.JournalInterval = ctx.GlobalDuration(utils.TxPoolJournalIntervalFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolPriceLimitFlag.Name) {
		cfg.PriceLimit = ctx.GlobalUint64(utils.TxPoolPriceLimitFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolPriceBumpFlag.Name) {
		cfg.PriceBump = ctx.GlobalUint64(utils.TxPoolPriceBumpFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolExecSlotsAccountFlag.Name) {
		cfg.ExecSlotsAccount = ctx.GlobalUint64(utils.TxPoolExecSlotsAccountFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolExecSlotsAllFlag.Name) {
		cfg.ExecSlotsAll = ctx.GlobalUint64(utils.TxPoolExecSlotsAllFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolNonExecSlotsAccountFlag.Name) {
		cfg.NonExecSlotsAccount = ctx.GlobalUint64(utils.TxPoolNonExecSlotsAccountFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolNonExecSlotsAllFlag.Name) {
		cfg.NonExecSlotsAll = ctx.GlobalUint64(utils.TxPoolNonExecSlotsAllFlag.Name)
	}

	cfg.KeepLocals = ctx.GlobalIsSet(utils.TxPoolKeepLocalsFlag.Name)

	if ctx.GlobalIsSet(utils.TxPoolLifetimeFlag.Name) {
		cfg.Lifetime = ctx.GlobalDuration(utils.TxPoolLifetimeFlag.Name)
	}

	// PN specific txpool setting
	if utils.NodeTypeFlag.Value == "pn" {
		cfg.EnableSpamThrottlerAtRuntime = !ctx.GlobalIsSet(utils.TxPoolSpamThrottlerDisableFlag.Name)
	}
}

// getNetworkID returns the associated network ID with whether or not the network is private.
func getNetworkId(ctx *cli.Context) (uint64, bool) {
	if ctx.GlobalIsSet(utils.BaobabFlag.Name) && ctx.GlobalIsSet(utils.CypressFlag.Name) {
		log.Fatalf("--baobab and --cypress must not be set together")
	}
	if ctx.GlobalIsSet(utils.BaobabFlag.Name) && ctx.GlobalIsSet(utils.NetworkIdFlag.Name) {
		log.Fatalf("--baobab and --networkid must not be set together")
	}
	if ctx.GlobalIsSet(utils.CypressFlag.Name) && ctx.GlobalIsSet(utils.NetworkIdFlag.Name) {
		log.Fatalf("--cypress and --networkid must not be set together")
	}

	switch {
	case ctx.GlobalIsSet(utils.CypressFlag.Name):
		logger.Info("Cypress network ID is set", "networkid", params.CypressNetworkId)
		return params.CypressNetworkId, false
	case ctx.GlobalIsSet(utils.BaobabFlag.Name):
		logger.Info("Baobab network ID is set", "networkid", params.BaobabNetworkId)
		return params.BaobabNetworkId, false
	case ctx.GlobalIsSet(utils.NetworkIdFlag.Name):
		networkId := ctx.GlobalUint64(utils.NetworkIdFlag.Name)
		logger.Info("A private network ID is set", "networkid", networkId)
		return networkId, true
	default:
		if utils.NodeTypeFlag.Value == "scn" || utils.NodeTypeFlag.Value == "spn" || utils.NodeTypeFlag.Value == "sen" {
			logger.Info("A Service Chain default network ID is set", "networkid", params.ServiceChainDefaultNetworkId)
			return params.ServiceChainDefaultNetworkId, true
		}
		logger.Info("Cypress network ID is set", "networkid", params.CypressNetworkId)
		return params.CypressNetworkId, false
	}
}

func setTxResendConfig(ctx *cli.Context, cfg *cn.Config) {
	// Set the Tx resending related configuration variables
	cfg.TxResendInterval = ctx.GlobalUint64(utils.TxResendIntervalFlag.Name)
	if cfg.TxResendInterval == 0 {
		cfg.TxResendInterval = cn.DefaultTxResendInterval
	}

	cfg.TxResendCount = ctx.GlobalInt(utils.TxResendCountFlag.Name)
	if cfg.TxResendCount < cn.DefaultMaxResendTxCount {
		cfg.TxResendCount = cn.DefaultMaxResendTxCount
	}
	cfg.TxResendUseLegacy = ctx.GlobalBool(utils.TxResendUseLegacyFlag.Name)
	logger.Debug("TxResend config", "Interval", cfg.TxResendInterval, "TxResendCount", cfg.TxResendCount, "UseLegacy", cfg.TxResendUseLegacy)
}

func (kCfg *klayConfig) setChainDataFetcherConfig(ctx *cli.Context) {
	cfg := &kCfg.ChainDataFetcher
	if ctx.GlobalBool(utils.EnableChainDataFetcherFlag.Name) {
		cfg.EnabledChainDataFetcher = true

		if ctx.GlobalIsSet(utils.ChainDataFetcherNoDefault.Name) {
			cfg.NoDefaultStart = true
		}
		if ctx.GlobalIsSet(utils.ChainDataFetcherNumHandlers.Name) {
			cfg.NumHandlers = ctx.GlobalInt(utils.ChainDataFetcherNumHandlers.Name)
		}
		if ctx.GlobalIsSet(utils.ChainDataFetcherJobChannelSize.Name) {
			cfg.JobChannelSize = ctx.GlobalInt(utils.ChainDataFetcherJobChannelSize.Name)
		}
		if ctx.GlobalIsSet(utils.ChainDataFetcherChainEventSizeFlag.Name) {
			cfg.BlockChannelSize = ctx.GlobalInt(utils.ChainDataFetcherChainEventSizeFlag.Name)
		}
		if ctx.GlobalIsSet(utils.ChainDataFetcherMaxProcessingDataSize.Name) {
			cfg.MaxProcessingDataSize = ctx.GlobalInt(utils.ChainDataFetcherMaxProcessingDataSize.Name)
		}

		mode := ctx.GlobalString(utils.ChainDataFetcherMode.Name)
		mode = strings.ToLower(mode)
		switch mode {
		case "kas": // kas option is not used.
			cfg.Mode = chaindatafetcher.ModeKAS
			cfg.KasConfig = makeKASConfig(ctx)
		case "kafka":
			cfg.Mode = chaindatafetcher.ModeKafka
			cfg.KafkaConfig = makeKafkaConfig(ctx)
		default:
			logger.Crit("unsupported chaindatafetcher mode (\"kas\", \"kafka\")", "mode", cfg.Mode)
		}
	}
}

// NOTE-klaytn
// Deprecated: KASConfig is not used anymore.
func checkKASDBConfigs(ctx *cli.Context) {
	if !ctx.GlobalIsSet(utils.ChainDataFetcherKASDBHostFlag.Name) {
		logger.Crit("DBHost must be set !", "key", utils.ChainDataFetcherKASDBHostFlag.Name)
	}
	if !ctx.GlobalIsSet(utils.ChainDataFetcherKASDBUserFlag.Name) {
		logger.Crit("DBUser must be set !", "key", utils.ChainDataFetcherKASDBUserFlag.Name)
	}
	if !ctx.GlobalIsSet(utils.ChainDataFetcherKASDBPasswordFlag.Name) {
		logger.Crit("DBPassword must be set !", "key", utils.ChainDataFetcherKASDBPasswordFlag.Name)
	}
	if !ctx.GlobalIsSet(utils.ChainDataFetcherKASDBNameFlag.Name) {
		logger.Crit("DBName must be set !", "key", utils.ChainDataFetcherKASDBNameFlag.Name)
	}
}

// NOTE-klaytn
// Deprecated: KASConfig is not used anymore.
func checkKASCacheInvalidationConfigs(ctx *cli.Context) {
	if !ctx.GlobalIsSet(utils.ChainDataFetcherKASCacheURLFlag.Name) {
		logger.Crit("The cache invalidation url is not set")
	}
	if !ctx.GlobalIsSet(utils.ChainDataFetcherKASBasicAuthParamFlag.Name) {
		logger.Crit("The authorization is not set")
	}
	if !ctx.GlobalIsSet(utils.ChainDataFetcherKASXChainIdFlag.Name) {
		logger.Crit("The x-chain-id is not set")
	}
}

// NOTE-klaytn
// Deprecated: KASConfig is not used anymore.
func makeKASConfig(ctx *cli.Context) *kas.KASConfig {
	kasConfig := kas.DefaultKASConfig

	checkKASDBConfigs(ctx)
	kasConfig.DBHost = ctx.GlobalString(utils.ChainDataFetcherKASDBHostFlag.Name)
	kasConfig.DBPort = ctx.GlobalString(utils.ChainDataFetcherKASDBPortFlag.Name)
	kasConfig.DBUser = ctx.GlobalString(utils.ChainDataFetcherKASDBUserFlag.Name)
	kasConfig.DBPassword = ctx.GlobalString(utils.ChainDataFetcherKASDBPasswordFlag.Name)
	kasConfig.DBName = ctx.GlobalString(utils.ChainDataFetcherKASDBNameFlag.Name)

	if ctx.GlobalBool(utils.ChainDataFetcherKASCacheUse.Name) {
		checkKASCacheInvalidationConfigs(ctx)
		kasConfig.CacheUse = true
		kasConfig.CacheInvalidationURL = ctx.GlobalString(utils.ChainDataFetcherKASCacheURLFlag.Name)
		kasConfig.BasicAuthParam = ctx.GlobalString(utils.ChainDataFetcherKASBasicAuthParamFlag.Name)
		kasConfig.XChainId = ctx.GlobalString(utils.ChainDataFetcherKASXChainIdFlag.Name)
	}
	return kasConfig
}

func makeKafkaConfig(ctx *cli.Context) *kafka.KafkaConfig {
	kafkaConfig := kafka.GetDefaultKafkaConfig()
	if ctx.GlobalIsSet(utils.ChainDataFetcherKafkaBrokersFlag.Name) {
		kafkaConfig.Brokers = ctx.GlobalStringSlice(utils.ChainDataFetcherKafkaBrokersFlag.Name)
	} else {
		logger.Crit("The kafka brokers must be set")
	}
	kafkaConfig.TopicEnvironmentName = ctx.GlobalString(utils.ChainDataFetcherKafkaTopicEnvironmentFlag.Name)
	kafkaConfig.TopicResourceName = ctx.GlobalString(utils.ChainDataFetcherKafkaTopicResourceFlag.Name)
	kafkaConfig.Partitions = int32(ctx.GlobalInt64(utils.ChainDataFetcherKafkaPartitionsFlag.Name))
	kafkaConfig.Replicas = int16(ctx.GlobalInt64(utils.ChainDataFetcherKafkaReplicasFlag.Name))
	kafkaConfig.SaramaConfig.Producer.MaxMessageBytes = ctx.GlobalInt(utils.ChainDataFetcherKafkaMaxMessageBytesFlag.Name)
	kafkaConfig.SegmentSizeBytes = ctx.GlobalInt(utils.ChainDataFetcherKafkaSegmentSizeBytesFlag.Name)
	kafkaConfig.MsgVersion = ctx.GlobalString(utils.ChainDataFetcherKafkaMessageVersionFlag.Name)
	kafkaConfig.ProducerId = ctx.GlobalString(utils.ChainDataFetcherKafkaProducerIdFlag.Name)
	requiredAcks := sarama.RequiredAcks(ctx.GlobalInt(utils.ChainDataFetcherKafkaRequiredAcksFlag.Name))
	if requiredAcks != sarama.NoResponse && requiredAcks != sarama.WaitForLocal && requiredAcks != sarama.WaitForAll {
		logger.Crit("not supported requiredAcks. it must be NoResponse(0), WaitForLocal(1), or WaitForAll(-1)", "given", requiredAcks)
	}
	kafkaConfig.SaramaConfig.Producer.RequiredAcks = requiredAcks
	return kafkaConfig
}

func (kCfg *klayConfig) setDBSyncerConfig(ctx *cli.Context) {
	cfg := &kCfg.DB
	if ctx.GlobalBool(utils.EnableDBSyncerFlag.Name) {
		cfg.EnabledDBSyncer = true

		if ctx.GlobalIsSet(utils.DBHostFlag.Name) {
			dbhost := ctx.GlobalString(utils.DBHostFlag.Name)
			cfg.DBHost = dbhost
		} else {
			logger.Crit("DBHost must be set !", "key", utils.DBHostFlag.Name)
		}
		if ctx.GlobalIsSet(utils.DBPortFlag.Name) {
			dbports := ctx.GlobalString(utils.DBPortFlag.Name)
			cfg.DBPort = dbports
		}
		if ctx.GlobalIsSet(utils.DBUserFlag.Name) {
			dbuser := ctx.GlobalString(utils.DBUserFlag.Name)
			cfg.DBUser = dbuser
		} else {
			logger.Crit("DBUser must be set !", "key", utils.DBUserFlag.Name)
		}
		if ctx.GlobalIsSet(utils.DBPasswordFlag.Name) {
			dbpasswd := ctx.GlobalString(utils.DBPasswordFlag.Name)
			cfg.DBPassword = dbpasswd
		} else {
			logger.Crit("DBPassword must be set !", "key", utils.DBPasswordFlag.Name)
		}
		if ctx.GlobalIsSet(utils.DBNameFlag.Name) {
			dbname := ctx.GlobalString(utils.DBNameFlag.Name)
			cfg.DBName = dbname
		} else {
			logger.Crit("DBName must be set !", "key", utils.DBNameFlag.Name)
		}
		if ctx.GlobalBool(utils.EnabledLogModeFlag.Name) {
			cfg.EnabledLogMode = true
		}
		if ctx.GlobalIsSet(utils.MaxIdleConnsFlag.Name) {
			cfg.MaxIdleConns = ctx.GlobalInt(utils.MaxIdleConnsFlag.Name)
		}
		if ctx.GlobalIsSet(utils.MaxOpenConnsFlag.Name) {
			cfg.MaxOpenConns = ctx.GlobalInt(utils.MaxOpenConnsFlag.Name)
		}
		if ctx.GlobalIsSet(utils.ConnMaxLifeTimeFlag.Name) {
			cfg.ConnMaxLifetime = ctx.GlobalDuration(utils.ConnMaxLifeTimeFlag.Name)
		}
		if ctx.GlobalIsSet(utils.DBSyncerModeFlag.Name) {
			cfg.Mode = strings.ToLower(ctx.GlobalString(utils.DBSyncerModeFlag.Name))
		}
		if ctx.GlobalIsSet(utils.GenQueryThreadFlag.Name) {
			cfg.GenQueryThread = ctx.GlobalInt(utils.GenQueryThreadFlag.Name)
		}
		if ctx.GlobalIsSet(utils.InsertThreadFlag.Name) {
			cfg.InsertThread = ctx.GlobalInt(utils.InsertThreadFlag.Name)
		}
		if ctx.GlobalIsSet(utils.BulkInsertSizeFlag.Name) {
			cfg.BulkInsertSize = ctx.GlobalInt(utils.BulkInsertSizeFlag.Name)
		}
		if ctx.GlobalIsSet(utils.EventModeFlag.Name) {
			cfg.EventMode = strings.ToLower(ctx.GlobalString(utils.EventModeFlag.Name))
		}
		if ctx.GlobalIsSet(utils.MaxBlockDiffFlag.Name) {
			cfg.MaxBlockDiff = ctx.GlobalUint64(utils.MaxBlockDiffFlag.Name)
		}
		if ctx.GlobalIsSet(utils.BlockSyncChannelSizeFlag.Name) {
			cfg.BlockChannelSize = ctx.GlobalInt(utils.BlockSyncChannelSizeFlag.Name)
		}
	}
}

func (kCfg *klayConfig) setServiceChainConfig(ctx *cli.Context) {
	cfg := &kCfg.ServiceChain

	// bridge service
	if ctx.GlobalBool(utils.MainBridgeFlag.Name) {
		cfg.EnabledMainBridge = true
		cfg.MainBridgePort = fmt.Sprintf(":%d", ctx.GlobalInt(utils.MainBridgeListenPortFlag.Name))
	}

	if ctx.GlobalBool(utils.SubBridgeFlag.Name) {
		cfg.EnabledSubBridge = true
		cfg.SubBridgePort = fmt.Sprintf(":%d", ctx.GlobalInt(utils.SubBridgeListenPortFlag.Name))
	}

	cfg.Anchoring = ctx.GlobalBool(utils.ServiceChainAnchoringFlag.Name)
	cfg.ChildChainIndexing = ctx.GlobalIsSet(utils.ChildChainIndexingFlag.Name)
	cfg.AnchoringPeriod = ctx.GlobalUint64(utils.AnchoringPeriodFlag.Name)
	cfg.SentChainTxsLimit = ctx.GlobalUint64(utils.SentChainTxsLimit.Name)
	cfg.ParentChainID = ctx.GlobalUint64(utils.ParentChainIDFlag.Name)
	cfg.VTRecovery = ctx.GlobalBool(utils.VTRecoveryFlag.Name)
	cfg.VTRecoveryInterval = ctx.GlobalUint64(utils.VTRecoveryIntervalFlag.Name)
	cfg.ServiceChainConsensus = utils.ServiceChainConsensusFlag.Value
	cfg.ServiceChainParentOperatorGasLimit = ctx.GlobalUint64(utils.ServiceChainParentOperatorTxGasLimitFlag.Name)
	cfg.ServiceChainChildOperatorGasLimit = ctx.GlobalUint64(utils.ServiceChainChildOperatorTxGasLimitFlag.Name)

	cfg.KASAnchor = ctx.GlobalBool(utils.KASServiceChainAnchorFlag.Name)
	if cfg.KASAnchor {
		cfg.KASAnchorPeriod = ctx.GlobalUint64(utils.KASServiceChainAnchorPeriodFlag.Name)
		if cfg.KASAnchorPeriod == 0 {
			cfg.KASAnchorPeriod = 1
			logger.Warn("KAS anchor period is set by 1")
		}

		cfg.KASAnchorUrl = ctx.GlobalString(utils.KASServiceChainAnchorUrlFlag.Name)
		if cfg.KASAnchorUrl == "" {
			logger.Crit("KAS anchor url should be set", "key", utils.KASServiceChainAnchorUrlFlag.Name)
		}

		cfg.KASAnchorOperator = ctx.GlobalString(utils.KASServiceChainAnchorOperatorFlag.Name)
		if cfg.KASAnchorOperator == "" {
			logger.Crit("KAS anchor operator should be set", "key", utils.KASServiceChainAnchorOperatorFlag.Name)
		}

		cfg.KASAccessKey = ctx.GlobalString(utils.KASServiceChainAccessKeyFlag.Name)
		if cfg.KASAccessKey == "" {
			logger.Crit("KAS access key should be set", "key", utils.KASServiceChainAccessKeyFlag.Name)
		}

		cfg.KASSecretKey = ctx.GlobalString(utils.KASServiceChainSecretKeyFlag.Name)
		if cfg.KASSecretKey == "" {
			logger.Crit("KAS secret key should be set", "key", utils.KASServiceChainSecretKeyFlag.Name)
		}

		cfg.KASXChainId = ctx.GlobalString(utils.KASServiceChainXChainIdFlag.Name)
		if cfg.KASXChainId == "" {
			logger.Crit("KAS x-chain-id should be set", "key", utils.KASServiceChainXChainIdFlag.Name)
		}

		cfg.KASAnchorRequestTimeout = ctx.GlobalDuration(utils.KASServiceChainAnchorRequestTimeoutFlag.Name)
	}

	cfg.DataDir = kCfg.Node.DataDir
	cfg.Name = kCfg.Node.Name
}

func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.CN.Genesis != nil {
		cfg.CN.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}
	io.WriteString(os.Stdout, comment)
	os.Stdout.Write(out)
	return nil
}
