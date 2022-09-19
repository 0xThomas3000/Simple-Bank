package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store defines all functions to run db queries individually and their combination within a transaction
type Store struct {
	// For individual Queries, we already have Queries struct, but each query only does 1 operation on 1 specific table
	// => Queries struct doesn't support Transaction => so, have to extend its functionality by embedding it inside
	// the Store struct like below(called a Composition: a preferred way to extend struct functionality instead of Inheritance)
	db       *sql.DB // All individual query functions provided by Queries'll be available to Store => can support TX by adding more funcs to that new struct
	*Queries         // In order to do above, Store needs to have sql.DB obj cuz it's required to create a new db TX
}

// NewStore() to create a new store obj
func NewStore(db *sql.DB) *Store {
	return &Store{ // Just build a new store obj and return it
		db:      db,      // db is the input sql.DB
		Queries: New(db), // Queries is created by calling the New() with that db object, New is created by sqlc and it'll return a Queries obj
	}
}

// To execute a generic database transaction
// This func is unexported cuz it starts with a lowercase letter => don't want external pkg call it directly, will provide an exported func for each specific TX instead
// - Takes a context, and a callback function as input. Then it'll start a new db TX
// In sum: it creates a new Queries obj with that TX, and call the callback function with the created Queries
//
//	and finally commit or rollback the TX based on the error returned by that function.
func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	// &sql.TxOptions{}: optional, allows us to set a custom isolation level for this TX
	// if we don't set it explicitly, then the default isolation level of the DB Server will be used (= read-committed in case of Postgres)
	// tx, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	tx, err := store.db.BeginTx(ctx, nil) // To start a new transaction, nil to use default value. BeginTx returns a TX obj/error
	if err != nil {
		return err
	}

	q := New(tx)    // Instead of passing in sql.DB, now pass in sql.Tx object (this works cuz the New() accepts a DBTX interface)
	err = fn(q)     // Now we have the Queries that runs within TX => we call the input function with that query and get back an error
	if err != nil { // Rollback the TX if we have an error, also return rollback error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr) // Report 2 errors if we also have Rollback Error => combine them into 1 Error to return
		}
		return err // if the Rollback is successful, return the original transaction error
	}

	return tx.Commit() // If all operations in TX are successful, commit TX and retuns its error to the Caller.
}

// TransferTxParams contains all necessary input parameters to transfer money between 2 accounts
type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// TransferTxResult contains the result of the transfer transaction
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`     // The created transfer record
	FromAccount Account  `json:"from_account"` // FromAccount after its balance is updated
	ToAccount   Account  `json:"to_account"`   // ToAccount after its balance is updated
	FromEntry   Entry    `json:"from_entry"`   // The entry of the FromAccount which records that money is moving out
	ToEntry     Entry    `json:"to_entry"`     // The entry of the ToAccount which records that money is moving in
}

// Later, will use this Key to get the Tx name from the input context of the TransferTx()
var txKey = struct{}{} // The second bracket (creating a "new empty obj of that type")

// TransferTx performs a money transfer from one account to the other.
// It creates a transfer record, add new account entries, update accounts' balance within a single database transaction
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult // Create an empty result

	err := store.execTx(ctx, func(q *Queries) error { // To create and run a new DB TX, pass in context and callback function
		/* Step 1: Create a transfer record */
		var err error

		txName := ctx.Value(txKey) // Get back the Tx name from Context

		// To implement this callback function: can use Queries obj to call any individual CRUD function that it provides
		// (this Queries obj is created from 1 single DB TX, so all of its provided methods we call will be run within that TX)
		// The output transfer will be save to result.Transfer
		fmt.Println(txName, "create transfer") // Print out the Tx name and the first operation: "create transfer"
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})

		// We're accessing the 'result' variable of the outer function(similar to the 'arg' variable)
		// => this makes the callback function become a closure (cuz Go lacks support for Generics type, Closure is often used if
		// want to get the result from callback() cuz the callback function doesn't know the exact type of the result it should return
		if err != nil {
			return err
		}

		/* Step 2: Add 2 Accounts entry, 1 for the FromAccount, 1 for the ToAccount */
		fmt.Println(txName, "create entry 1")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		fmt.Println(txName, "create entry 2")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		/********* Step 3: update accounts' balance *********/
		// Moving the Money out of the fromAccount
		fmt.Println(txName, "get account 1")
		account1, err := q.GetAccount(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}

		fmt.Println(txName, "update account 1")
		result.FromAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.FromAccountID,
			Balance: account1.Balance - arg.Amount,
		})
		if err != nil {
			return err
		}

		// Do similar thing to move those money into the toAccount
		fmt.Println(txName, "get account 2")
		account2, err := q.GetAccount(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}

		fmt.Println(txName, "update account 2")
		result.ToAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
			ID:      arg.ToAccountID,
			Balance: account2.Balance + arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err // Returns the result and error of the execTx() call
}
