package data

import (
	"log"
	"testing"

	"github.com/lacker/coinkit/consensus"
)

func TestInsertAndGet(t *testing.T) {
	db := NewTestDatabase(0)
	qs, _ := consensus.MakeTestQuorumSlice(4)
	block := &Block{
		Slot:  1,
		Chunk: NewEmptyChunk(),
		C:     7,
		H:     8,
		D:     qs,
	}
	err := db.InsertBlock(block)
	if err != nil {
		t.Fatal(err)
	}
	db.Commit()
	if db.GetBlock(4) != nil {
		t.Fatal("block should be nonexistent")
	}
	b2 := db.GetBlock(1)
	if b2.C != block.C {
		t.Fatalf("block changed: %+v -> %+v", block, b2)
	}
	if b2.Chunk == nil {
		t.Fatalf("block chunk was nil on retrieval")
	}
	if b2.D == nil {
		t.Fatalf("block quorum slice was nil on retrieval")
	}
}

func TestCantInsertTwice(t *testing.T) {
	db := NewTestDatabase(0)
	block := &Block{
		Slot:  1,
		Chunk: NewEmptyChunk(),
		C:     1,
		H:     2,
	}
	err := db.InsertBlock(block)
	if err != nil {
		t.Fatal(err)
	}
	err = db.InsertBlock(block)
	if err == nil {
		t.Fatal("a block should not save twice")
	}

	if db.LastBlock() != nil {
		t.Fatal("insert should not have worked without commit")
	}
	db.Rollback()
	err = db.InsertBlock(block)
	if err != nil {
		t.Fatal(err)
	}
	db.Commit()
	if db.LastBlock() == nil {
		t.Fatal("insert should work after rollback")
	}
}

func TestLastBlock(t *testing.T) {
	db := NewTestDatabase(0)
	b := db.LastBlock()
	if b != nil {
		t.Fatalf("expected last block nil but got %+v", b)
	}
	b = &Block{
		Slot:  1,
		Chunk: NewEmptyChunk(),
	}
	err := db.InsertBlock(b)
	if err != nil {
		t.Fatal(err)
	}

	// Before commit, the insert should not be visible
	b2 := db.LastBlock()
	if b2 != nil {
		t.Fatalf("expected b2 nil but got: %+v", b2)
	}
	db.Commit()
	b.Slot = 2
	err = db.InsertBlock(b)
	if err != nil {
		t.Fatal(err)
	}
	db.Commit()
	b3 := db.LastBlock()
	if b3.Slot != b.Slot {
		t.Fatalf("b3: %+v", b3)
	}

	// We should also be able to retrieve it with a query message
	qm := &QueryMessage{
		Block: b.Slot,
	}
	dm := db.HandleQueryMessage(qm)
	if dm == nil {
		t.Fatalf("got nil data message")
	}
	b4 := dm.Blocks[b.Slot]
	if b4 == nil || b4.Slot != b.Slot {
		t.Fatalf("got bad data message: %+v", dm)
	}
}

func TestForBlocks(t *testing.T) {
	db := NewTestDatabase(0)
	for i := 1; i <= 5; i++ {
		b := &Block{
			Slot:  i,
			Chunk: NewEmptyChunk(),
			C:     7,
		}
		if db.InsertBlock(b) != nil {
			t.Fatal("block could not save")
		}
		db.Commit()
	}
	count := db.ForBlocks(func(b *Block) {
		if b.C != 7 {
			t.Fatal("expected C = 7")
		}
	})
	if count != 5 {
		t.Fatal("expected count = 5")
	}
	log.Print(db.TotalSizeInfo())
}

func TestGetDocuments(t *testing.T) {
	db := NewTestDatabase(0)
	for a := 1; a <= 2; a++ {
		for b := 1; b <= 2; b++ {
			d := NewDocument(uint64(10*a+b), map[string]interface{}{
				"a": a,
				"b": b,
			})
			err := db.InsertDocument(d)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	docs := db.GetDocuments(map[string]interface{}{"a": 2, "b": 1}, 2)
	if len(docs) != 0 {
		t.Fatal("expected no docs visible before commit")
	}
	db.Commit()
	docs = db.GetDocuments(map[string]interface{}{"a": 2, "b": 1}, 2)
	if len(docs) != 1 {
		t.Fatalf("expected one doc but got: %+v", docs)
	}
}

func TestGetDocumentsNoResults(t *testing.T) {
	db := NewTestDatabase(0)
	docs := db.GetDocuments(map[string]interface{}{"blorp": "hi"}, 3)
	if len(docs) != 0 {
		t.Fatalf("expected zero docs but got: %+v", docs)
	}
}

const benchmarkMax = 400

func databaseForBenchmarking() *Database {
	db := NewTestDatabase(0)
	log.Printf("populating db for benchmarking")
	items := 0
	for a := 0; a < benchmarkMax; a++ {
		if a != 0 && a%10 == 0 {
			log.Printf("inserted %d items", items)
		}
		for b := 0; b < benchmarkMax; b++ {
			c := b*benchmarkMax + a + 1
			d := NewDocument(uint64(c), map[string]interface{}{
				"a": a,
				"b": b,
				"c": c,
			})
			err := db.InsertDocument(d)
			if err != nil {
				log.Fatal(err)
			}
			items++
		}
	}
	log.Printf("database is populated with %d items", items)
	return db
}

func BenchmarkOneConstraint(b *testing.B) {
	db := databaseForBenchmarking()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := i%(benchmarkMax*benchmarkMax) + 1
		docs := db.GetDocuments(map[string]interface{}{"c": c}, 2)
		if len(docs) != 1 {
			log.Fatalf("expected one doc for c = %d but got: %+v", c, docs)
		}
	}
}

func BenchmarkTwoConstraints(b *testing.B) {
	db := databaseForBenchmarking()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := i % benchmarkMax
		b := ((i - a) / benchmarkMax) % benchmarkMax
		docs := db.GetDocuments(map[string]interface{}{"a": a, "b": b}, 2)
		if len(docs) != 1 {
			log.Fatalf("expected one doc but got: %+v", docs)
		}
	}
}

func TestMaxBalance(t *testing.T) {
	db := NewTestDatabase(0)
	mb := db.MaxBalance()
	if mb != 0 {
		t.Fatalf("got max balance %d but expected 0", mb)
	}

	a := &Account{
		Owner:    "alex",
		Sequence: 1,
		Balance:  10,
	}
	b := &Account{
		Owner:    "bob",
		Sequence: 2,
		Balance:  5,
	}
	db.UpsertAccount(a)
	db.UpsertAccount(b)
	mb = db.MaxBalance()
	if mb != 0 {
		t.Fatalf("got max balance %d before commit, but expected 0", mb)
	}

	db.Commit()
	mb = db.MaxBalance()
	if mb != 10 {
		t.Fatalf("got max balance %d", mb)
	}
}

func TestAccounts(t *testing.T) {
	db := NewTestDatabase(0)
	if db.GetAccount("bob") != nil {
		t.Fatalf("db should be empty")
	}
	nothing := func(a *Account) {}
	if db.ForAccounts(nothing) != 0 {
		t.Fatalf("ForAccounts on empty db should be 0")
	}
	a := &Account{
		Owner:    "bob",
		Sequence: 3,
		Balance:  4,
	}
	db.UpsertAccount(a)
	db.Commit()
	if db.GetAccount("bob") == nil {
		t.Fatalf("bob should exist now")
	}
	numAccounts := db.ForAccounts(nothing)
	if numAccounts != 1 {
		t.Fatalf("there should be 1 thing in the db now, but there was %d", numAccounts)
	}
	a.Owner = "bob2"
	db.UpsertAccount(a)
	db.Commit()
	if db.ForAccounts(nothing) != 2 {
		t.Fatalf("there should be 2 things in the db now")
	}
	m := &QueryMessage{
		Account: "bob",
	}
	dm := db.HandleQueryMessage(m)
	if dm == nil || dm.I != 0 || dm.Accounts["bob"].Balance != 4 {
		t.Fatalf("got unexpected data message: %+v", dm)
	}
}
