# Meta Marketing API Integration

This document outlines the requirements for integrating with the Meta Marketing API. The goal is to create a system capable of programmatically creating and managing advertisements on the Facebook platform.
Official API Documentation: https://developers.facebook.com/docs/marketing-apis

## Key Requirements:

1.  **SDK Usage:**
    *   Utilize the Go SDK: https://github.com/justwatch/facebook-marketing-api-golang-sdk for all interactions with the Meta Marketing API.

2.  **Ad Creation Workflow:**
    *   The system must support the creation of a complete ad campaign structure using a pre-configured Ad Account. This involves creating and linking the following entities in the specified order:
        1.  **Ad Account:** The top-level account (selected from managed accounts) under which all advertising activities are managed.
        2.  **Campaign:** Defines a specific advertising objective (e.g., brand awareness, conversions).
        3.  **Ad Set (Ad Group):** Defines the targeting, budget, schedule, bidding, and placement for a group of ads.
        4.  **Ad Creative:** The actual content that users see (e.g., image, video, text). This links to an Ad Set.
        5.  **Ad:** The specific media and copy for the ad. This is associated with an Ad.

2.1. **Ad Account Management:**
    *   **Dedicated Storage:** Ad Account information must be stored in a dedicated database table (e.g., `ad_accounts`). This table will serve as a central repository for all ad accounts across different platforms.
    *   **Key Information:** Each record in the `ad_accounts` table should include, but not be limited to:
        *   `id`: Primary key for internal system reference.
        *   `platform_account_id`: The actual ID of the account on the advertising platform (e.g., Meta Ad Account ID `act_xxxxxxxx`).
        *   `name`: A user-friendly name for the account (e.g., "Client X - Meta US").
        *   `platform_type`: Enum/String indicating the platform (e.g., `META`, `GOOGLE_ADS`, `TIKTOK_ADS`).
        *   `credentials`: Securely stored authentication details (e.g., encrypted access tokens, API keys, refresh tokens). Consider using a secrets management system.
        *   `status`: Indicates the current state of the account's integration (e.g., `ACTIVE`, `INACTIVE`, `REQUIRES_REAUTH`, `SUSPENDED`).
        *   `owner_user_id` or `team_id`: (Optional) To associate the ad account with an internal user or team.
        *   `additional_config`: A JSONB or text field for platform-specific settings or metadata not covered by common fields.
    *   **Provisioning:** Ad accounts are assumed to be pre-configured in the system (e.g., added manually via an admin interface or a separate provisioning process). The system is not responsible for creating new ad accounts on the Meta platform itself, but rather for using existing ones.
    *   **Selection in API:** The ad creation API (see point 3) must accept an identifier (e.g., our internal `id` or the `platform_account_id`) to specify which Ad Account should be used for the new ad campaign. The system will then retrieve the necessary details (like access tokens) from the `ad_accounts` table.

3.  **API Endpoint for Ad Creation:**
    *   Implement a dedicated API endpoint (e.g., `POST /api/v1/ads/meta`) that accepts ad creation requests.
    *   This endpoint will receive all necessary parameters to define the ad components listed in point 2.
    *   The request payload should be well-defined and validated.

4.  **Database Integration and Asynchronous Processing:**
    *   **Data Persistence:** Before attempting to create an ad via the Meta API, all ad-related data (campaign details, ad set parameters, ad creative information, etc.) must be saved to a persistent database.
    *   **Status Tracking:** Each ad creation job stored in the database should have a status field. Suggested statuses:
        *   `PENDING`: The ad creation request has been received and stored but not yet processed.
        *   `PROCESSING`: The system is currently attempting to create the ad via the Meta API.
        *   `SUCCESS`: The ad was successfully created on the Meta platform. Store relevant IDs (e.g., Meta Campaign ID, Ad Set ID, Ad ID).
        *   `FAILED`: The ad creation failed. Store error messages or codes from the API for debugging.
        *   `PARTIAL_SUCCESS`: (Optional) If some components were created but others failed.
    *   **Asynchronous Creation:** Implement a mechanism (e.g., a background worker polling the database, or a message queue system) to process `PENDING` ad creation jobs. This decouples the API request from the actual Meta API interaction, improving API responsiveness and resilience.

5.  **Extensibility for Multi-Platform Support:**
    *   **Generic Data Models:** Design database schemas with future extensibility in mind. While the initial focus is Meta, the system may later support other advertising platforms (e.g., Google Ads, TikTok Ads).
    *   The `ad_accounts` table (detailed in 2.1) is a prime example of this, centralizing account information with a `platform_type` discriminator.
    *   Consider abstracting common advertising concepts (e.g., campaigns, ad groups/sets, ads, creatives) into generic tables/models, with platform-specific details in separate, related tables or JSONB fields.
    *   This will facilitate easier integration of new platforms without requiring major schema overhauls. For example:
        *   A `campaigns` table with common fields (name, objective_type, status).
        *   A `meta_campaign_details` table linked to `campaigns` for Meta-specific attributes.
        *   A `google_campaign_details` table linked to `campaigns` for Google-specific attributes.

## Next Steps for AI-Assisted Development:

The following steps will guide the AI in implementing the Meta Marketing API integration:

1.  **Define Ad Account Management Database Schema & SQLC Models (in `/models`):**
    *   **Table Structure:** Create the SQL schema for the `ad_accounts` table as detailed in section "2.1. Ad Account Management."
        *   Ensure columns like `id` (primary key, e.g., UUID), `platform_account_id` (TEXT, indexed), `name` (TEXT), `platform_type` (TEXT or ENUM: 'META', 'GOOGLE_ADS', etc.), `credentials` (TEXT, to store encrypted tokens), `status` (TEXT or ENUM: 'ACTIVE', 'INACTIVE', etc.), `owner_user_id` (UUID, optional, foreign key if applicable), `additional_config` (JSONB), `created_at` (TIMESTAMPTZ), `updated_at` (TIMESTAMPTZ).
    *   **SQLC Queries:** Write SQL queries for CRUD operations (Create, Read, Update, Delete) for the `ad_accounts` table.
    *   **SQLC Generation:** Run `sqlc generate` to create the Go model and query code in `models/gen/`.

2.  **Implement Ad Account Management API Endpoints (in `/routes` and `/tests`):**
    *   **Route Definitions:** In a new file (e.g., `routes/ad_accounts.go`), define HTTP handlers for managing ad accounts.
        *   `POST /api/v1/ad_accounts`: Create a new ad account. Request body should match the `ad_accounts` table structure (excluding auto-generated fields like `id`, `created_at`, `updated_at`).
        *   `GET /api/v1/ad_accounts`: List all ad accounts (with pagination).
        *   `GET /api/v1/ad_accounts/{accountId}`: Get a specific ad account by its internal `id`.
        *   `PUT /api/v1/ad_accounts/{accountId}`: Update an existing ad account.
        *   `DELETE /api/v1/ad_accounts/{accountId}`: Delete an ad account (consider soft delete).
    *   **Handler Logic:** Implement the handler functions using the `handlerFunc(ctx context.Context, req RequestType) (ResponseType, error)` signature. These handlers will use the SQLC-generated code to interact with the database.
    *   **Request/Response Structs:** Define Go structs for API request and response payloads.
    *   **Registration:** Register these new routes in `routes.RegisterRoutes` within `cmd/mojito/main.go`.
    *   **Authentication/Authorization:** Ensure appropriate middleware (e.g., `RequireAuth` from `middleware/handler.go`) is applied to these routes.
    *   **API Tests:** Create Hurl tests in the `/tests` directory (e.g., `tests/ad_accounts.hurl`) to cover all new ad account management endpoints.

3.  **Define Ad Campaign & Related Entities Database Schema & SQLC Models (in `/models`):**
    *   **Generic Campaign Table (`campaign_jobs` or `ad_submissions`):**
        *   This table will store the initial request for creating an ad, as per section "4. Database Integration and Asynchronous Processing."
        *   Columns: `id` (PK), `ad_account_id` (FK to `ad_accounts.id`), `platform_type` (TEXT, e.g., 'META'), `job_status` (TEXT: 'PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'PARTIAL_SUCCESS'), `payload` (JSONB, storing the full ad creation request), `error_message` (TEXT, nullable), `platform_campaign_id` (TEXT, nullable, stores Meta Campaign ID upon success), `created_at`, `updated_at`.
    *   **(Optional) Detailed Platform-Specific Tables (for extensibility, as per section 5):**
        *   `meta_campaigns`: `id` (PK), `campaign_job_id` (FK), `meta_campaign_id` (TEXT, unique), `name`, `objective`, `status_on_platform`, etc.
        *   `meta_ad_sets`: `id` (PK), `meta_campaign_id` (FK), `meta_ad_set_id` (TEXT, unique), `name`, `targeting_criteria` (JSONB), `budget`, `schedule`, etc.
        *   `meta_ads`: `id` (PK), `meta_ad_set_id` (FK), `meta_ad_id` (TEXT, unique), `name`, `creative_id` (FK to `meta_ad_creatives`), etc.
        *   `meta_ad_creatives`: `id` (PK), `meta_creative_id` (TEXT, unique), `name`, `body_text`, `image_hash`, `video_id`, `call_to_action`, etc.
        *   *Consider if these detailed tables are needed immediately or if storing platform IDs in the `campaign_jobs` table and relying on the `payload` for details is sufficient for the initial Meta integration.*
    *   **SQLC Queries:** Write SQL queries for managing these tables, focusing on:
        *   Creating new `campaign_jobs`.
        *   Fetching `PENDING` jobs for processing.
        *   Updating job status and storing platform IDs or error messages.
    *   **SQLC Generation:** Run `sqlc generate`.

4.  **Implement Ad Creation API Endpoint & Asynchronous Processing Logic (in `/routes`, background worker, and `/tests`):**
    *   **API Endpoint (`POST /api/v1/ads/meta` as per section 3):**
        *   In a new file (e.g., `routes/ads_meta.go`), define the handler for this endpoint.
        *   **Request Payload:** Define a comprehensive request struct that includes all necessary information to create a Meta Campaign, Ad Set, Ad, and Ad Creative. This will include the `ad_account_id` (our internal ID) to be used.
        *   **Handler Logic:**
            1.  Validate the request payload.
            2.  Retrieve the specified Ad Account details (especially credentials) from the `ad_accounts` table.
            3.  Store the ad creation request (the payload) into the `campaign_jobs` table with `status: PENDING`.
            4.  Return an immediate acknowledgment to the client (e.g., the `campaign_job_id` and `status: PENDING`).
    *   **Asynchronous Worker (New Package, e.g., `/workers/ad_processor.go`):**
        *   Implement a background worker (e.g., using goroutines and a ticker, or a message queue if available).
        *   The worker will periodically query the `campaign_jobs` table for `PENDING` jobs for `platform_type: META`.
        *   For each job:
            1.  Update job status to `PROCESSING`.
            2.  Parse the `payload`.
            3.  Retrieve Ad Account credentials.
            4.  Initialize the Meta Marketing API SDK (`facebook-marketing-api-golang-sdk`).
            5.  Sequentially create Campaign, Ad Set, Ad Creative, and Ad using the SDK.
            6.  Handle API errors gracefully.
            7.  Update the `campaign_jobs` table with `SUCCESS` and platform IDs, or `FAILED` and error messages. If using detailed tables, populate them as well.
    *   **API Tests:** Create Hurl tests in `/tests` (e.g., `tests/ads_meta.hurl`) for the `POST /api/v1/ads/meta` endpoint. These tests will primarily verify the job submission and initial `PENDING` status. Testing the full asynchronous flow might require more complex integration tests or manual verification.

5.  **Identify and Map Key SDK Functions:**
    *   Thoroughly review the `facebook-marketing-api-golang-sdk` documentation and source code.
    *   Pinpoint the specific functions and their required parameters for:
        *   Authenticating with an Ad Account.
        *   Creating a Campaign (e.g., `Campaign.Create()`).
        *   Creating an Ad Set (e.g., `AdSet.Create()`).
        *   Creating an Ad Creative (e.g., `AdCreative.Create()`).
        *   Creating an Ad (e.g., `Ad.Create()`).
    *   Document these mappings for use in the asynchronous worker.

6.  **Configuration for Meta API:**
    *   Ensure `common/config.go` and `config/config.yaml.example` are updated to handle any necessary global configurations for the Meta API integration, if any (e.g., API version, default timeouts), though most credentials will be per-ad-account.