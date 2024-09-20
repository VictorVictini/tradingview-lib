package tradingview

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

/*
Searches for symbols using a search term and type.
If the search term is blank then it will get everything.
*/
func (search *Search) SearchSymbols(term string, searchType SearchType) ([]interface{}, error) {
	search.mutex.Lock()
	defer search.mutex.Unlock()

	// make the string have no starting/leading spaces + have it uppercase
	term = strings.ToUpper(strings.TrimSpace(term))

	// check if it is a valid search term
	if !regexp.MustCompile(`^[A-Z0-9 :]*$`).MatchString(term) {
		return nil, errors.New("search term must be either empty or only have characters ['A-Z', '0-9', ' ', ':']")
	}

	// reset search bool if searching again
	if search.hasSearched {
		search.hasSearched = false
	}

	// send the search request (for the first page)
	resp, err := sendRawSearchRequest(term, searchType, 1)
	if err != nil {
		return nil, err
	}
	// close after this function is finished with the response
	defer resp.Body.Close()

	// parse it from JSON
	res, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// get how many remaining symbols are left (symbols amount/50 + 1 = page count)
	remaining, ok := res["symbols_remaining"].(float64)
	if !ok {
		return nil, errors.New("unexpected response format: symbols_remaining field missing or incorrect type")
	}

	// set the search results
	symbols, ok := res["symbols"].([]interface{})
	if !ok {
		return nil, errors.New("unexpected response format: symbols field missing or incorrect type")
	}

	// initialise properties
	search.currPage = 1
	search.maxPages = int(math.Max(1.0, math.Ceil(remaining/SEARCH_PAGE_SIZE)+1)) // +1 to include first page

	search.currTerm = term
	search.currSearchType = searchType
	search.hasSearched = true

	return symbols, nil
}

/*
Move onto the next page of the search.
You must call Search() first.
*/
func (search *Search) NextPage() ([]interface{}, error) {
	search.mutex.Lock()
	defer search.mutex.Unlock()

	// user hasn't searched for anything yet, return error
	if !search.hasSearched {
		return nil, errors.New("you must first call Search()")
	}

	// stop, we're at the max page limit
	if search.currPage == search.maxPages {
		return nil, errors.New("reached max pages")
	}

	// send a search request for the next page
	resp, err := sendRawSearchRequest(search.currTerm, search.currSearchType, search.currPage+1)
	if err != nil {
		return nil, err
	}
	// close after this function is finished with the response
	defer resp.Body.Close()

	// parse it from JSON
	res, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// set the search results
	symbols, ok := res["symbols"].([]interface{})
	if !ok {
		return nil, errors.New("unexpected response format: symbols field missing or incorrect type")
	}

	// increment the page number now
	search.currPage++
	return symbols, nil
}

/*
Move onto the previous page of the search.
You must call Search() first.
*/
func (search *Search) PrevPage() ([]interface{}, error) {
	search.mutex.Lock()
	defer search.mutex.Unlock()

	// user hasn't searched for anything yet, return error
	if !search.hasSearched {
		return nil, errors.New("you must first call Search()")
	}

	// stop, we're at the min page limit
	if search.currPage == 1 {
		return nil, errors.New("cannot go back any more pages")
	}

	// send a search request for the previous page
	resp, err := sendRawSearchRequest(search.currTerm, search.currSearchType, search.currPage-1)
	if err != nil {
		return nil, err
	}
	// close after this function is finished with the response
	defer resp.Body.Close()

	// parse it from JSON
	res, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// set the search results
	symbols, ok := res["symbols"].([]interface{})
	if !ok {
		return nil, errors.New("unexpected response format: symbols field missing or incorrect type")
	}

	// decrement the page number now
	search.currPage--
	return symbols, nil
}

/*
Get the current page number in the search.
You must call Search() first.
*/
func (search *Search) GetCurrentPageNo() (int, error) {
	search.mutex.Lock()
	defer search.mutex.Unlock()

	// user hasn't searched for anything yet, return error
	if !search.hasSearched {
		return 0, errors.New("you must first call Search()")
	}

	return search.currPage, nil
}

/*
Get the max number of pages that are available in the current search.
You must call Search() first.
*/
func (search *Search) GetMaxPages() (int, error) {
	search.mutex.Lock()
	defer search.mutex.Unlock()

	// user hasn't searched for anything yet, return error
	if !search.hasSearched {
		return 0, errors.New("you must first call Search()")
	}

	return search.maxPages, nil
}

/*
Have you searched yet?
*/
func (search *Search) HasSearched() bool {
	search.mutex.Lock()
	defer search.mutex.Unlock()

	return search.hasSearched
}

/*
Raw search get request, used internally.
*/
func sendRawSearchRequest(term string, searchType SearchType, page int) (*http.Response, error) {
	var req *http.Request
	var err error

	// separate exchange and rest of symbol
	separate := strings.SplitN(term, ":", 2)
	exchange := ""

	if len(separate) == 2 {
		// exchange is the first part before the colon
		exchange = separate[0]

		// escape the rest of the term so that it can be used inside the GET url (e.g. ":" -> %3A, " " -> +)
		term = url.QueryEscape(separate[1])
	} else {
		// escape the entire term if there is no exchange specified
		term = url.QueryEscape(term)
	}

	// create the base url using search term and type
	baseUrl := "https://symbol-search.tradingview.com/symbol_search/v3/?text=" + term + "&exchange=" + exchange + "&search_type=" + string(searchType)
	fmt.Println(baseUrl)

	// decrement since page numbers are 0-indexed
	page--

	// if it is not the first page then we have to multiply the page number by 50 to get the remaining symbols
	if page > 0 {
		req, err = http.NewRequest(http.MethodGet, baseUrl+"&start="+strconv.Itoa(page*SEARCH_PAGE_SIZE), nil)
	} else {
		// the first page, we don't specify the page size
		req, err = http.NewRequest(http.MethodGet, baseUrl, nil)
	}

	// return if error
	if err != nil {
		return nil, err
	}

	// set origin header (needed)
	req.Header.Set("Origin", TV_ORIGIN_URL)

	// send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// check if the server accepted it as a valid request
	if resp.StatusCode != 200 {
		return nil, errors.New("request was sent but server gave an error")
	}

	// return the response itself (the body will need to be closed manually)
	return resp, nil
}

/*
Parse a http response into JSON, used internally.
*/
func parseResponse(resp *http.Response) (map[string]interface{}, error) {
	// parse the response body into a byte array
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// the result map itself
	var res map[string]interface{}

	// use the byte array to read as JSON and assign it to the res var
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	// return the result
	return res, nil
}
