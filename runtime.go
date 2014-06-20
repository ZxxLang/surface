// +build !race
package surface

// Via /pkg/runtime/runtime.h
type _Lock struct {
	key uintptr
}

// Via /pkg/runtime/chan.c
type _SudoG struct {
	_           *uintptr // g and selgen constitute
	selgen      uint32   // a weak pointer to g
	link        *_SudoG
	releasetime int64
	elem        *byte // data element
}

type _WaitQ struct {
	first *_SudoG
	last  *_SudoG
}

// Via /pkg/runtime/chan.c
type _Hchan struct {
	qcount   uint
	dataqsiz uint
	elemsize uint16
	pad      uint16
	closed   bool
	alg      *uintptr
	sendx    uint
	recvx    uint
	recvq    _WaitQ
	sendq    _WaitQ
	_Lock
}

// Via /pkg/runtime/hashmap.c
type _Hmap struct {
	count      uint
	flags      uint32
	hash0      uint32
	B          uint8
	keysize    uint8
	valuesize  uint8
	bucketsize uint8
	buckets    *byte
	oldbuckets *byte
	nevacuate  uintptr
}

func maplen(m IWord) int {
	if m == nil {
		return 0
	}
	p := (*_Hmap)(m)
	return int(p.count)
}

// !race
// void
// reflect·maplen(Hmap *h, intgo len)
// {
// 	if(h == nil)
// 		len = 0;
// 	else {
// 		len = h->count;
// 		if(raceenabled)
// 			runtime·racereadpc(h, runtime·getcallerpc(&h), reflect·maplen);
// 	}
// 	FLUSH(&len);
// }
func chancap(ch IWord) int {
	if ch == nil {
		return 0
	}
	p := (*_Hchan)(ch)
	return int(p.dataqsiz)
}

func chanlen(ch IWord) int {
	if ch == nil {
		return 0
	}
	p := (*_Hchan)(ch)
	return int(p.qcount)
}

/**
func    mapaccess(t *rtype, m iword, key iword) (val iword, ok bool)
runtime·mapaccess(MapType *t, Hmap *h, byte *ak, byte *av, bool *pres)
{
	byte *res;
	Type *elem;

	elem = t->elem;
	if(h == nil || h->count == 0) {
		elem->alg->copy(elem->size, av, nil);
		*pres = false;
		return;
	}

	res = hash_lookup(t, h, &ak);

	if(res != nil) {
		*pres = true;
		elem->alg->copy(elem->size, av, res);
	} else {
		*pres = false;
		elem->alg->copy(elem->size, av, nil);
	}
}

static byte*
hash_lookup(MapType *t, Hmap *h, byte **keyp)
{
	void *key;
	uintptr hash;
	uintptr bucket, oldbucket;
	Bucket *b;
	uint8 top;
	uintptr i;
	bool eq;
	byte *k, *k2, *v;

	key = *keyp;
	if(docheck)
		check(t, h);
	if(h->count == 0)
		return nil;
	hash = h->hash0;
	t->key->alg->hash(&hash, t->key->size, key);
	bucket = hash & (((uintptr)1 << h->B) - 1);
	if(h->oldbuckets != nil) {
		oldbucket = bucket & (((uintptr)1 << (h->B - 1)) - 1);
		b = (Bucket*)(h->oldbuckets + oldbucket * h->bucketsize);
		if(evacuated(b)) {
			b = (Bucket*)(h->buckets + bucket * h->bucketsize);
		}
	} else {
		b = (Bucket*)(h->buckets + bucket * h->bucketsize);
	}
	top = hash >> (sizeof(uintptr)*8 - 8);
	if(top == 0)
		top = 1;
	do {
		for(i = 0, k = b->data, v = k + h->keysize * BUCKETSIZE; i < BUCKETSIZE; i++, k += h->keysize, v += h->valuesize) {
			if(b->tophash[i] == top) {
				k2 = IK(h, k);
				t->key->alg->equal(&eq, t->key->size, key, k2);
				if(eq) {
					*keyp = k2;
					return IV(h, v);
				}
			}
		}
		b = b->overflow;
	} while(b != nil);
	return nil;
}
#define IK(h, p) (((h)->flags & IndirectKey) != 0 ? *(byte**)(p) : (p))
#define IV(h, p) (((h)->flags & IndirectValue) != 0 ? *(byte**)(p) : (p))
typedef struct Bucket Bucket;
struct Bucket
{
	// Note: the format of the Bucket is encoded in ../../cmd/gc/reflect.c and
	// ../reflect/type.go.  Don't change this structure without also changing that code!
	uint8  tophash[BUCKETSIZE]; // top 8 bits of hash of each entry (0 = empty)
	Bucket *overflow;           // overflow bucket, if any
	byte   data[1];             // BUCKETSIZE keys followed by BUCKETSIZE values
};
*/

/**
struct	Alg
{
	void	(*hash)(uintptr*, uintptr, void*);
	void	(*equal)(bool*, uintptr, void*, void*);
	void	(*print)(uintptr, void*);
	void	(*copy)(uintptr, void*, void*);
};
*/
