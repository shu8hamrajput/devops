#!/usr/bin/env python3
"""
Script to populate the Splitwise database with mock data for testing.

This script creates:
- Multiple users
- Groups with users
- Expenses with splits
- User-to-user expenses

Requirements:
    pip3 install psycopg2-binary

Usage:
    python3 populate_mock_data.py

Or with environment variables:
    DB_HOST=localhost DB_PORT=5432 DB_NAME=splitwise DB_USER=postgres DB_PASSWORD=postgres python3 populate_mock_data.py
"""

import psycopg2
import uuid
import random
from datetime import datetime, timedelta
from decimal import Decimal

# Database configuration
# Can be overridden by environment variables
import os

DB_CONFIG = {
    'host': os.getenv('DB_HOST', 'localhost'),
    'port': int(os.getenv('DB_PORT', '5432')),
    'database': os.getenv('DB_NAME', 'splitwise'),
    'user': os.getenv('DB_USER', 'postgres'),
    'password': os.getenv('DB_PASSWORD', 'postgres')
}

def get_db_connection():
    """Create and return a database connection."""
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        return conn
    except Exception as e:
        print(f"Error connecting to database: {e}")
        print("Make sure PostgreSQL is running and the database exists.")
        return None

def create_users(conn, count=10):
    """Create mock users."""
    cursor = conn.cursor()
    users = []
    
    first_names = ['Alice', 'Bob', 'Charlie', 'Diana', 'Eve', 'Frank', 'Grace', 'Henry', 'Ivy', 'Jack']
    last_names = ['Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis', 'Rodriguez', 'Martinez']
    
    print(f"Creating {count} users...")
    for i in range(count):
        user_id = str(uuid.uuid4())
        name = f"{first_names[i % len(first_names)]} {last_names[i % len(last_names)]}"
        email = f"user{i+1}@example.com"
        password_hash = "$2a$10$dummyhash"  # Dummy hash for testing
        
        now = datetime.now()
        try:
            cursor.execute("""
                INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s, %s)
                ON CONFLICT (email) DO NOTHING
            """, (user_id, name, email, password_hash, now, now))
            users.append({'id': user_id, 'name': name, 'email': email})
        except Exception as e:
            print(f"Warning: Could not create user {name}: {e}")
    
    conn.commit()
    # Get all users (including existing ones)
    cursor.execute("SELECT id, name, email FROM users ORDER BY created_at LIMIT %s", (count,))
    all_users = cursor.fetchall()
    users = [{'id': str(u[0]), 'name': u[1], 'email': u[2]} for u in all_users]
    print(f"Total users available: {len(users)}")
    return users

def create_groups(conn, users, count=5):
    """Create mock groups with users."""
    cursor = conn.cursor()
    groups = []
    
    group_names = [
        'Weekend Trip',
        'House Rent',
        'Dinner Club',
        'Vacation Fund',
        'Grocery Sharing'
    ]
    
    print(f"Creating {count} groups...")
    for i in range(count):
        group_id = str(uuid.uuid4())
        name = group_names[i % len(group_names)] if i < len(group_names) else f"Group {i+1}"
        now = datetime.now()
        
        # Create group
        cursor.execute("""
            INSERT INTO groups (id, name, created_at, updated_at)
            VALUES (%s, %s, %s, %s)
        """, (group_id, name, now, now))
        
        # Add 3-5 random users to each group
        num_members = random.randint(3, min(5, len(users)))
        group_users = random.sample(users, num_members)
        
        for user in group_users:
            cursor.execute("""
                INSERT INTO group_users (group_id, user_id)
                VALUES (%s, %s)
                ON CONFLICT DO NOTHING
            """, (group_id, user['id']))
        
        groups.append({
            'id': group_id,
            'name': name,
            'user_ids': [u['id'] for u in group_users]
        })
    
    conn.commit()
    print(f"Created {len(groups)} groups")
    return groups

def create_expenses_with_splits(conn, groups, users):
    """Create mock expenses with splits."""
    cursor = conn.cursor()
    expenses = []
    
    expense_descriptions = [
        'Dinner at restaurant',
        'Uber ride',
        'Groceries',
        'Movie tickets',
        'Coffee',
        'Lunch',
        'Gas',
        'Hotel booking',
        'Concert tickets',
        'Shopping'
    ]
    
    print("Creating expenses with splits...")
    for group in groups:
        # Create 3-5 expenses per group
        num_expenses = random.randint(3, 5)
        
        for _ in range(num_expenses):
            expense_id = str(uuid.uuid4())
            description = random.choice(expense_descriptions)
            amount = round(random.uniform(20.0, 500.0), 2)
            
            # Random user from group pays
            paid_by = random.choice(group['user_ids'])
            now = datetime.now()
            
            # Create expense
            cursor.execute("""
                INSERT INTO expenses (id, description, amount, paid_by, group_id, owed_by, created_at, updated_at, deleted_at)
                VALUES (%s, %s, %s, %s, %s, NULL, %s, %s, NULL)
            """, (expense_id, description, amount, paid_by, group['id'], now, now))
            
            # Create equal splits for all group members
            num_members = len(group['user_ids'])
            amount_per_person = round(amount / num_members, 2)
            
            # Adjust last person's amount to account for rounding
            total_distributed = 0.0
            for idx, user_id in enumerate(group['user_ids']):
                if idx == len(group['user_ids']) - 1:
                    # Last person gets the remainder
                    split_amount = round(amount - total_distributed, 2)
                else:
                    split_amount = amount_per_person
                    total_distributed += split_amount
                
                split_id = str(uuid.uuid4())
                cursor.execute("""
                    INSERT INTO expense_splits (id, expense_id, user_id, amount, share_type, share_value, created_at, updated_at)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                """, (split_id, expense_id, user_id, split_amount, 'EQUAL', 1.0, now, now))
            
            expenses.append({
                'id': expense_id,
                'description': description,
                'amount': amount,
                'group_id': group['id']
            })
    
    conn.commit()
    print(f"Created {len(expenses)} expenses with splits")
    return expenses

def create_user_to_user_expenses(conn, users, count=5):
    """Create user-to-user expenses."""
    cursor = conn.cursor()
    expenses = []
    
    descriptions = [
        'Lunch payment',
        'Coffee',
        'Shared taxi',
        'Movie ticket',
        'Book purchase'
    ]
    
    print(f"Creating {count} user-to-user expenses...")
    for _ in range(count):
        expense_id = str(uuid.uuid4())
        description = random.choice(descriptions)
        amount = round(random.uniform(10.0, 100.0), 2)
        
        # Pick two different users
        user1, user2 = random.sample(users, 2)
        paid_by = user1['id']
        owed_by = user2['id']
        now = datetime.now()
        
        try:
            cursor.execute("""
                INSERT INTO expenses (id, description, amount, paid_by, group_id, owed_by, created_at, updated_at, deleted_at)
                VALUES (%s, %s, %s, %s, NULL, %s, %s, %s, NULL)
            """, (expense_id, description, Decimal(str(amount)), paid_by, owed_by, now, now))
        except Exception as e:
            print(f"Warning: Could not create user-to-user expense {description}: {e}")
            conn.rollback()
            continue
        
        expenses.append({
            'id': expense_id,
            'description': description,
            'amount': amount,
            'paid_by': paid_by,
            'owed_by': owed_by
        })
    
    conn.commit()
    print(f"Created {len(expenses)} user-to-user expenses")
    return expenses

def main():
    """Main function to populate database."""
    print("=" * 60)
    print("Splitwise Mock Data Population Script")
    print("=" * 60)
    print(f"\nConnecting to database: {DB_CONFIG['host']}:{DB_CONFIG['port']}/{DB_CONFIG['database']}")
    
    conn = get_db_connection()
    if not conn:
        print("\nFailed to connect to database.")
        print("Make sure PostgreSQL is running and accessible.")
        print("You can also set environment variables:")
        print("  DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD")
        return
    
    try:
        # Create users
        users = create_users(conn, count=10)
        if not users or len(users) == 0:
            print("No users found or created. Exiting.")
            return
        
        # Create groups
        groups = create_groups(conn, users, count=5)
        
        # Create expenses with splits
        expenses = create_expenses_with_splits(conn, groups, users)
        
        # Create user-to-user expenses
        user_expenses = create_user_to_user_expenses(conn, users, count=5)
        
        print("\n" + "=" * 60)
        print("Summary:")
        print(f"  Users: {len(users)}")
        print(f"  Groups: {len(groups)}")
        print(f"  Group Expenses: {len(expenses)}")
        print(f"  User-to-User Expenses: {len(user_expenses)}")
        print("=" * 60)
        print("\nMock data populated successfully!")
        print("\nYou can now test the application with this data.")
        print("\nSample user credentials:")
        print("  Email: user1@example.com")
        print("  (Note: You'll need to register/login through the app)")
        
    except Exception as e:
        print(f"\nError: {e}")
        import traceback
        traceback.print_exc()
        if conn:
            conn.rollback()
    finally:
        if conn:
            conn.close()

if __name__ == '__main__':
    main()

