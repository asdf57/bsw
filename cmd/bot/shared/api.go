package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	apimodels "github.com/asdf57/bsw/internal/models/api"
)

func apiURL() (string, error) {
	url := strings.TrimSpace(os.Getenv("API_URL"))
	if url == "" {
		return "", fmt.Errorf("missing required env var: API_URL")
	}
	return url, nil
}

func GetPaymentResponses() ([]apimodels.PaymentResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(apiURL + "/api/v1/payment/all")
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var payments []apimodels.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payments); err != nil {
		return nil, fmt.Errorf("error decoding payment response: %w", err)
	}

	return payments, nil
}

func GetPaymentResponsesByTags(tags []string, op string) ([]apimodels.PaymentResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	for _, tag := range tags {
		cleaned := strings.ToLower(strings.TrimSpace(tag))
		if cleaned != "" {
			values.Add("tags", cleaned)
		}
	}
	values.Set("op", strings.ToLower(strings.TrimSpace(op)))

	resp, err := http.Get(apiURL + "/api/v1/payment/tags?" + values.Encode())
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var payments []apimodels.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payments); err != nil {
		return nil, fmt.Errorf("error decoding payment response: %w", err)
	}

	return payments, nil
}

func GetPaymentResponse(paymentID uint) (*apimodels.PaymentResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/payment/%d", apiURL, paymentID))
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var payment apimodels.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		return nil, fmt.Errorf("error decoding payment response: %w", err)
	}

	return &payment, nil
}

func GetTags() ([]apimodels.TagResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(apiURL + "/api/v1/tags")
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var tags []apimodels.TagResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("error decoding tag response: %w", err)
	}

	return tags, nil
}

func CreateTag(name string) (*apimodels.TagResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	payload := apimodels.Tag{Name: name}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tag payload: %w", err)
	}

	resp, err := http.Post(apiURL+"/api/v1/tags", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var tag apimodels.TagResponse
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, fmt.Errorf("error decoding tag response: %w", err)
	}

	return &tag, nil
}

func CreatePayment(payment *apimodels.Payment) error {
	apiURL, err := apiURL()
	if err != nil {
		return err
	}

	body, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshall payment json")
	}

	resp, err := http.Post(apiURL+"/api/v1/payment", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create payment")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return nil
}

func DeletePayment(paymentID uint) error {
	apiURL, err := apiURL()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/payment/%d", apiURL, paymentID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing payment deletion: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return nil
}

func UpdatePayment(paymentID uint, payment *apimodels.Payment) (*apimodels.PaymentResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment json: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/payment/%d", apiURL, paymentID), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating update request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing payment update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var updated apimodels.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, fmt.Errorf("error decoding updated payment response: %w", err)
	}

	return &updated, nil
}

func SettleDebts(owedBy string, owedTo string) (*apimodels.SettleDebtsResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	payload := apimodels.SettleDebtsRequest{OwedBy: owedBy, OwedTo: owedTo}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settle payload: %w", err)
	}

	resp, err := http.Post(apiURL+"/api/v1/debts/settle", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to settle debts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var settleResp apimodels.SettleDebtsResponse
	if err := json.NewDecoder(resp.Body).Decode(&settleResp); err != nil {
		return nil, fmt.Errorf("error decoding settle response: %w", err)
	}

	return &settleResp, nil
}

func GetDebtResponses(currency string) ([]apimodels.DebtResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("currency", NormalizePaymentCurrency(currency))

	resp, err := http.Get(apiURL + "/api/v1/debts?" + values.Encode())
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var debts []apimodels.DebtResponse
	if err := json.NewDecoder(resp.Body).Decode(&debts); err != nil {
		return nil, fmt.Errorf("error decoding debt response: %w", err)
	}

	return debts, nil
}

func GetSettlements() ([]apimodels.SettlementResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(apiURL + "/api/v1/debts/settlements")
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var settlements []apimodels.SettlementResponse
	if err := json.NewDecoder(resp.Body).Decode(&settlements); err != nil {
		return nil, fmt.Errorf("error decoding settlement response: %w", err)
	}

	return settlements, nil
}

func ReverseSettlement(settlementID uint) (*apimodels.SettlementResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/v1/debts/settlements/%d/reverse", apiURL, settlementID), "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse settlement: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var settlement apimodels.SettlementResponse
	if err := json.NewDecoder(resp.Body).Decode(&settlement); err != nil {
		return nil, fmt.Errorf("error decoding settlement response: %w", err)
	}

	return &settlement, nil
}

func GetUserStats(user string, currency string) (*apimodels.UserStatsResponse, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Set("user", strings.TrimSpace(user))
	values.Set("currency", NormalizePaymentCurrency(currency))

	resp, err := http.Get(apiURL + "/api/v1/stats/user?" + values.Encode())
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	var stats apimodels.UserStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("error decoding stats response: %w", err)
	}

	return &stats, nil
}

func GetUsers() ([]apimodels.UserSummary, error) {
	apiURL, err := apiURL()
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(apiURL + "/api/v1/user")
	if err != nil {
		return nil, fmt.Errorf("error making http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	var users []apimodels.UserSummary
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("error decoding user response: %w", err)
	}

	return users, nil
}

func CreateUser(username string, discordHandle string) error {
	apiURL, err := apiURL()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/user", apiURL)
	payload := apimodels.User{Name: username, DiscordHandle: discordHandle}
	marshaledJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal user payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(marshaledJSON))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d %s: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return nil
}

func NormalizePaymentCurrency(currency string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(currency))
	if cleaned == "" {
		cleaned = "USD"
	}
	return cleaned
}
