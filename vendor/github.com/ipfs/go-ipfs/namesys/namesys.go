package namesys

import (
	"context"
	"strings"
	"time"

	opts "github.com/ipfs/go-ipfs/namesys/opts"
	path "github.com/ipfs/go-ipfs/path"

	mh "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	lru "gx/ipfs/QmVYxfoJQiZijTgPNHCHgHELvQpbsJNTg6Crmc3dQkj3yy/golang-lru"
	routing "gx/ipfs/QmZ383TySJVeZWzGnWui6pRcKyYZk9VkKTuW7tmKRWk5au/go-libp2p-routing"
	isd "gx/ipfs/QmZmmuAXgX73UQmX1jRKjTGmjzq24Jinqkq8vzkBtno4uX/go-is-domain"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	ci "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	ds "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore"
)

// mpns (a multi-protocol NameSystem) implements generic IPFS naming.
//
// Uses several Resolvers:
// (a) IPFS routing naming: SFS-like PKI names.
// (b) dns domains: resolves using links in DNS TXT records
// (c) proquints: interprets string as the raw byte data.
//
// It can only publish to: (a) IPFS routing naming.
//
type mpns struct {
	dnsResolver, proquintResolver, ipnsResolver resolver
	ipnsPublisher                               Publisher

	cache *lru.Cache
}

// NewNameSystem will construct the IPFS naming system based on Routing
func NewNameSystem(r routing.ValueStore, ds ds.Datastore, cachesize int) NameSystem {
	var cache *lru.Cache
	if cachesize > 0 {
		cache, _ = lru.New(cachesize)
	}

	return &mpns{
		dnsResolver:      NewDNSResolver(),
		proquintResolver: new(ProquintResolver),
		ipnsResolver:     NewIpnsResolver(r),
		ipnsPublisher:    NewIpnsPublisher(r, ds),
		cache:            cache,
	}
}

const DefaultResolverCacheTTL = time.Minute

// Resolve implements Resolver.
func (ns *mpns) Resolve(ctx context.Context, name string, options ...opts.ResolveOpt) (path.Path, error) {
	if strings.HasPrefix(name, "/ipfs/") {
		return path.ParsePath(name)
	}

	if !strings.HasPrefix(name, "/") {
		return path.ParsePath("/ipfs/" + name)
	}

	return resolve(ctx, ns, name, opts.ProcessOpts(options), "/ipns/")
}

// resolveOnce implements resolver.
func (ns *mpns) resolveOnce(ctx context.Context, name string, options *opts.ResolveOpts) (path.Path, time.Duration, error) {
	if !strings.HasPrefix(name, "/ipns/") {
		name = "/ipns/" + name
	}
	segments := strings.SplitN(name, "/", 4)
	if len(segments) < 3 || segments[0] != "" {
		log.Debugf("invalid name syntax for %s", name)
		return "", 0, ErrResolveFailed
	}

	key := segments[2]

	p, ok := ns.cacheGet(key)
	var err error
	if !ok {
		// Resolver selection:
		// 1. if it is a multihash resolve through "ipns".
		// 2. if it is a domain name, resolve through "dns"
		// 3. otherwise resolve through the "proquint" resolver
		var res resolver
		if _, err := mh.FromB58String(key); err == nil {
			res = ns.ipnsResolver
		} else if isd.IsDomain(key) {
			res = ns.dnsResolver
		} else {
			res = ns.proquintResolver
		}

		var ttl time.Duration
		p, ttl, err = res.resolveOnce(ctx, key, options)
		if err != nil {
			return "", 0, ErrResolveFailed
		}
		ns.cacheSet(key, p, ttl)
	}

	if len(segments) > 3 {
		p, err = path.FromSegments("", strings.TrimRight(p.String(), "/"), segments[3])
	}
	return p, 0, err
}

// Publish implements Publisher
func (ns *mpns) Publish(ctx context.Context, name ci.PrivKey, value path.Path) error {
	return ns.PublishWithEOL(ctx, name, value, time.Now().Add(DefaultRecordTTL))
}

func (ns *mpns) PublishWithEOL(ctx context.Context, name ci.PrivKey, value path.Path, eol time.Time) error {
	id, err := peer.IDFromPrivateKey(name)
	if err != nil {
		return err
	}
	if err := ns.ipnsPublisher.PublishWithEOL(ctx, name, value, eol); err != nil {
		return err
	}
	ttl := DefaultResolverCacheTTL
	if ttEol := eol.Sub(time.Now()); ttEol < ttl {
		ttl = ttEol
	}
	ns.cacheSet(peer.IDB58Encode(id), value, ttl)
	return nil
}
