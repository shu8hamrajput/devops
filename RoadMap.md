# Splitwise App - Sprint Improvement Plan

## Current State Assessment

### ✅ What's Working
- Authentication system (JWT with refresh tokens)
- User management with PostgreSQL persistence
- Basic expense and group creation
- Frontend UI with React
- Docker setup with PostgreSQL, Redis, Jaeger
- Hexagonal architecture (Ports & Adapters)
- Distributed tracing with Jaeger

### ❌ Critical Gaps
- **Expenses and Groups stored in memory** - data lost on restart
- **No expense splitting logic** - expenses not split among group members
- **No balance calculation** - can't see who owes what
- **No settlement mechanism**
- **Missing GET /groups endpoint**
- **No expense editing/deletion**
- **No expense splitting options** (equal, percentage, custom)

---

## Sprint Plan (4 Sprints)

### 🚀 Sprint 1: Data Persistence & Core Features (2 weeks)
**Goal:** Persist critical data and add missing core functionality

#### Backend Tasks
- [x] **PostgreSQL repositories for Expenses**
  - Create `postgres_expense_repository.go`
  - Add expense table schema with migrations
  - Implement all ExpenseRepository methods
  - Update `main.go` to use PostgreSQL instead of memory

- [x] **PostgreSQL repositories for Groups**
  - Create `postgres_group_repository.go`
  - Add group and group_users junction table
  - Implement all GroupRepository methods
  - Update `main.go` to use PostgreSQL instead of memory

- [x] **Add GET /groups endpoint**
  - Add `GetAllGroups` method to GroupService
  - Add handler for `GET /groups`
  - Add frontend API integration

- [x] **Expense CRUD operations**
  - Add `UpdateExpense` to ExpenseService
  - Add `DeleteExpense` to ExpenseService
  - Add handlers: `PUT /expenses/{id}`
  - Add handlers: `DELETE /expenses/{id}` (soft delete)
  - Add validation and authorization checks

- [x] **Group CRUD operations**
  - Add `UpdateGroup` to GroupService
  - Add `DeleteGroup` to GroupService
  - Add `RemoveUserFromGroup` to GroupService
  - Add handlers: `PUT /groups/{id}`, `DELETE /groups/{id}`, `DELETE /groups/{id}/users/{user_id}`

#### Frontend Tasks
- [x] **Expense management UI**
  - Add edit expense functionality
  - Add delete expense with confirmation
  - Improve expense list display with actions

- [x] **Group management UI**
  - Add edit group functionality
  - Add delete group with confirmation
  - Add remove user from group
  - Display all groups in Groups component

- [x] **Error handling improvements**
  - Better error messages
  - Toast notifications for success/error
  - Loading states for all async operations

#### Testing
- [x] Unit tests for new repository methods
- [x] Integration tests for CRUD endpoints
- [ ] Frontend component tests (Optional - can be done in Sprint 4)

**Deliverables:**
- All data persisted in PostgreSQL
- Full CRUD for expenses and groups
- No data loss on server restart

---

### 💰 Sprint 2: Expense Splitting & Balance Calculation (2 weeks)
**Goal:** Implement core Splitwise functionality - splitting expenses and calculating balances

#### Backend Tasks
- [ ] **Expense splitting entity**
  - Add `ExpenseSplit` entity with fields: ExpenseID, UserID, Amount, ShareType
  - Add split types: EQUAL, PERCENTAGE, EXACT_AMOUNT, SHARES
  - Create migration for expense_splits table

- [ ] **Expense splitting service**
  - Add `SplitExpense` method to ExpenseService
  - Implement equal split logic
  - Implement percentage split logic
  - Implement custom amount split logic
  - Validate splits sum to expense amount

- [ ] **Balance calculation service**
  - Create new `BalanceService` with port interface
  - Implement `CalculateGroupBalance(groupID)` - returns who owes whom
  - Implement `CalculateUserBalance(userID)` - returns net balance
  - Implement `CalculateUserToUserBalance(userID1, userID2)` - returns balance between two users
  - Use debt simplification algorithm (minimize transactions)

- [ ] **Balance endpoints**
  - `GET /groups/{id}/balance` - group balance summary
  - `GET /users/{id}/balance` - user's overall balance
  - `GET /users/{id}/balance/{other_user_id}` - balance with specific user

- [ ] **Update expense creation**
  - Modify `CreateExpense` to accept split configuration
  - Auto-split expenses equally among group members by default
  - Store splits in database

#### Frontend Tasks
- [ ] **Expense splitting UI**
  - Add split type selector (Equal, Percentage, Custom)
  - Add split configuration form
  - Display split details in expense list
  - Show who owes what for each expense

- [ ] **Balance visualization**
  - Add balance card to Dashboard
  - Show "You owe" and "You are owed" amounts
  - Display group balances in Groups component
  - Add balance view in Expenses component

- [ ] **User-to-user balance view**
  - Show balance between current user and other users
  - Display in Dashboard and Users component

#### Testing
- [ ] Unit tests for splitting logic
- [ ] Unit tests for balance calculation
- [ ] Integration tests for balance endpoints
- [ ] Edge case testing (negative balances, zero amounts)

**Deliverables:**
- Expenses can be split among group members
- Balance calculation working correctly
- Users can see who owes what

---

### 🎯 Sprint 3: Settlement & Advanced Features (2 weeks)
**Goal:** Add settlement mechanism and enhance user experience

#### Backend Tasks
- [ ] **Settlement entity and service**
  - Create `Settlement` entity (ID, FromUserID, ToUserID, Amount, GroupID, Status, CreatedAt)
  - Create `SettlementService` with port interface
  - Implement `CreateSettlement` - record a payment
  - Implement `GetSettlementsByUser` - get user's settlement history
  - Implement `GetSettlementsByGroup` - get group's settlement history

- [ ] **Settlement endpoints**
  - `POST /settlements` - create a settlement
  - `GET /users/{id}/settlements` - get user settlements
  - `GET /groups/{id}/settlements` - get group settlements
  - `PUT /settlements/{id}` - update settlement status
  - `DELETE /settlements/{id}` - cancel pending settlement

- [ ] **Balance simplification**
  - Implement debt simplification algorithm
  - Minimize number of transactions needed
  - Add endpoint `GET /groups/{id}/simplified-balance` - shows minimal transactions

- [ ] **Expense categories**
  - Add `Category` field to Expense entity
  - Add category enum/table
  - Filter expenses by category

- [ ] **Expense search and filtering**
  - Add search by description
  - Filter by date range
  - Filter by amount range
  - Filter by category
  - Add pagination to expense lists

#### Frontend Tasks
- [ ] **Settlement UI**
  - Add "Settle Up" button in balance views
  - Settlement form (amount, to user, optional note)
  - Settlement history view
  - Mark settlement as paid

- [ ] **Expense filtering and search**
  - Add search bar in Expenses component
  - Add date range picker
  - Add category filter
  - Add amount range filter
  - Add pagination controls

- [ ] **Simplified balance view**
  - Show minimal transactions needed
  - Visual representation of debt chain
  - "Settle all" functionality

- [ ] **Dashboard enhancements**
  - Add expense statistics (total spent, total owed, total owed to you)
  - Add recent expenses widget
  - Add upcoming settlements reminder

#### Testing
- [ ] Settlement service tests
- [ ] Debt simplification algorithm tests
- [ ] Filtering and search tests

**Deliverables:**
- Users can settle debts
- Simplified balance view
- Enhanced filtering and search

---

### 🚀 Sprint 4: Performance, Security & Polish (2 weeks)
**Goal:** Optimize performance, enhance security, and polish the application

#### Backend Tasks
- [ ] **Caching improvements**
  - Cache balance calculations in Redis
  - Cache user groups list
  - Implement cache invalidation strategy
  - Add TTL for cached data

- [ ] **Database optimization**
  - Add indexes on frequently queried fields
  - Add database connection pooling
  - Query optimization for balance calculations
  - Add database query logging in dev mode

- [ ] **API improvements**
  - Add request validation middleware
  - Add rate limiting
  - Add request/response logging
  - Add API versioning (v1 prefix)

- [ ] **Security enhancements**
  - Add input sanitization
  - Add SQL injection prevention (parameterized queries)
  - Add CORS configuration for production
  - Add environment-based configuration
  - Secure JWT secret management

- [ ] **Error handling**
  - Standardized error responses
  - Error logging with context
  - Graceful error recovery

- [ ] **API documentation**
  - Add OpenAPI/Swagger documentation
  - Document all endpoints
  - Add request/response examples

#### Frontend Tasks
- [ ] **Performance optimization**
  - Implement React.memo for expensive components
  - Add lazy loading for routes
  - Optimize API calls (debouncing, caching)
  - Add loading skeletons instead of spinners

- [ ] **UI/UX improvements**
  - Improve responsive design
  - Add dark mode support
  - Improve accessibility (ARIA labels, keyboard navigation)
  - Add animations and transitions
  - Improve form validation and feedback

- [ ] **State management**
  - Consider adding React Context or Redux for global state
  - Optimize re-renders
  - Add optimistic updates for better UX

- [ ] **Error boundaries**
  - Add React error boundaries
  - Better error messages for users
  - Error reporting/logging

#### DevOps & Infrastructure
- [ ] **CI/CD pipeline**
  - Add GitHub Actions or similar
  - Automated testing on PR
  - Automated deployment
  - Docker image optimization

- [ ] **Monitoring and logging**
  - Add application metrics (Prometheus)
  - Add structured logging
  - Set up error tracking (Sentry or similar)
  - Add health check endpoints

- [ ] **Documentation**
  - Update README with setup instructions
  - Add API documentation
  - Add architecture diagrams
  - Add deployment guide

#### Testing
- [ ] **End-to-end tests**
  - Add E2E tests for critical flows
  - Test authentication flow
  - Test expense creation and splitting
  - Test settlement flow

- [ ] **Load testing**
  - Test API under load
  - Identify bottlenecks
  - Optimize slow queries

**Deliverables:**
- Optimized performance
- Enhanced security
- Production-ready application
- Comprehensive documentation

---

## Future Enhancements (Post-Sprint 4)

### Phase 2 Features
- [ ] **Multi-currency support**
  - Currency conversion
  - Exchange rate handling

- [ ] **Recurring expenses**
  - Set up recurring expenses
  - Automatic expense creation

- [ ] **Expense attachments**
  - Upload receipts/images
  - File storage integration

- [ ] **Notifications**
  - Email notifications
  - Push notifications
  - In-app notifications

- [ ] **Reports and analytics**
  - Spending reports
  - Category-wise breakdown
  - Export to CSV/PDF

- [ ] **Group features**
  - Group admins
  - Group settings
  - Group invitations

- [ ] **Mobile app**
  - React Native app
  - Mobile-optimized UI

---

## Success Metrics

### Sprint 1
- ✅ All data persisted (0% data loss on restart)
- ✅ 100% CRUD coverage for expenses and groups
- ✅ All endpoints return < 200ms response time

### Sprint 2
- ✅ Expense splitting working for all split types
- ✅ Balance calculation accurate
- ✅ Balance API response time < 500ms

### Sprint 3
- ✅ Settlement flow working end-to-end
- ✅ Users can filter expenses by 3+ criteria
- ✅ Simplified balance reduces transactions by 50%+

### Sprint 4
- ✅ API response time < 100ms (p95)
- ✅ Zero security vulnerabilities
- ✅ 90%+ test coverage
- ✅ Documentation complete

---

## Technical Debt to Address

1. **Memory repositories** - Replace with PostgreSQL (Sprint 1)
2. **Missing error handling** - Add comprehensive error handling (Sprint 4)
3. **No input validation** - Add validation middleware (Sprint 4)
4. **No pagination** - Add pagination to all list endpoints (Sprint 3)
5. **Hardcoded values** - Move to environment variables (Sprint 4)
6. **No tests** - Add comprehensive test suite (All sprints)
7. **CORS wildcard** - Configure proper CORS (Sprint 4)

---

## Notes

- Each sprint is designed to be 2 weeks long
- Prioritize Sprint 1 and Sprint 2 as they contain critical functionality
- Sprint 3 and Sprint 4 can be adjusted based on business priorities
- Consider breaking down larger tasks into smaller subtasks
- Regular code reviews and pair programming recommended
- Daily standups to track progress

