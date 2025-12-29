#!/bin/bash

# curl apis to create 20 users, 5 groups, and 100 expenses. all which form a interconnected graph of users, groups, and expenses.

set -e  # Exit on error

# Check dependencies
if ! command -v curl &> /dev/null; then
    echo "Error: curl is required but not installed."
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed. Install with: brew install jq"
    exit 1
fi

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "Error: Server is not running on http://localhost:8080"
    echo "Please start the server first with: go run main.go"
    exit 1
fi

echo "Creating 20 users..."
# create and store all user_ids in a variable
user_ids=()
for i in {1..20}; do
    response=$(curl -s -X POST http://localhost:8080/users \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"User $i\",\"email\":\"user$i@example.com\"}")
    user_id=$(echo $response | jq -r '.id')
    if [ "$user_id" == "null" ] || [ -z "$user_id" ]; then
        echo "  Error creating user $i: $response"
        exit 1
    fi
    user_ids+=($user_id)
    echo "  Created user $i: $user_id"
done

echo ""
echo "Creating 5 groups with randomly selected users..."
# create groups with randomly selected user_ids from user_ids variable
group_ids=()
group_names=("Weekend Trip" "Office Lunch" "House Party" "Road Trip" "Birthday Celebration")

for i in {0..4}; do
    # Randomly select 3-8 users for each group
    num_users=$((RANDOM % 6 + 3))
    selected_users=()
    
    # Shuffle and pick random users
    shuffled_users=($(printf '%s\n' "${user_ids[@]}" | shuf))
    for j in $(seq 0 $((num_users - 1))); do
        selected_users+=(${shuffled_users[$j]})
    done
    
    # Create JSON array of user IDs
    user_ids_json=$(printf '%s\n' "${selected_users[@]}" | jq -R . | jq -s .)
    
    response=$(curl -s -X POST http://localhost:8080/groups \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"${group_names[$i]}\",\"user_ids\":$user_ids_json}")
    group_id=$(echo $response | jq -r '.id')
    if [ "$group_id" == "null" ] || [ -z "$group_id" ]; then
        echo "  Error creating group ${group_names[$i]}: $response"
        exit 1
    fi
    group_ids+=($group_id)
    echo "  Created group ${group_names[$i]}: $group_id (with ${#selected_users[@]} users)"
done

echo ""
echo "Creating 100 expenses with randomly selected users and groups..."
# Temporarily disable exit on error for expense creation to allow partial completion
set +e
# create expenses with randomly selected user_ids and group_ids from user_ids and group_ids variables
expense_descriptions=(
    "Dinner at restaurant"
    "Uber ride"
    "Groceries"
    "Movie tickets"
    "Coffee"
    "Lunch"
    "Breakfast"
    "Hotel booking"
    "Gas"
    "Parking"
    "Concert tickets"
    "Bar tab"
    "Shopping"
    "Taxi"
    "Snacks"
    "Drinks"
    "Fast food"
    "Pizza"
    "Ice cream"
    "Bakery"
)

for i in {1..100}; do
    # Randomly select a user and group
    random_user_idx=$((RANDOM % ${#user_ids[@]}))
    random_group_idx=$((RANDOM % ${#group_ids[@]}))
    random_desc_idx=$((RANDOM % ${#expense_descriptions[@]}))
    
    paid_by=${user_ids[$random_user_idx]}
    group_id=${group_ids[$random_group_idx]}
    description="${expense_descriptions[$random_desc_idx]} #$i"
    amount=$(awk "BEGIN {printf \"%.2f\", ($RANDOM % 5000 + 100) / 100}")
    
    response=$(curl -s -X POST http://localhost:8080/expenses \
        -H "Content-Type: application/json" \
        -d "{\"description\":\"$description\",\"amount\":$amount,\"paid_by\":\"$paid_by\",\"group_id\":\"$group_id\"}")
    
    expense_id=$(echo $response | jq -r '.id')
    if [ "$expense_id" == "null" ] || [ -z "$expense_id" ]; then
        echo "  Error creating expense $i: $response"
        # Continue instead of exit to allow partial completion
        continue
    fi
    
    if [ $((i % 10)) -eq 0 ]; then
        echo "  Created expense $i/100: $expense_id"
    fi
done

set -e  # Re-enable exit on error

echo ""
echo "Setup complete!"
echo "  - Users: ${#user_ids[@]}"
echo "  - Groups: ${#group_ids[@]}"
echo "  - Expenses: 100"