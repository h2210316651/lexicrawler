package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/log"
	"github.com/go-shiori/go-readability" 
)

// CrawlerConfig, CrawledData, Crawler, NewCrawler, Crawl, getCachedData, cacheData, fetchDynamicContent, captureScreenshot, generateMarkdown, parseSrcset, resolveURL, applyHeuristics - remain the same

// CrawlerConfig holds configuration for the crawler
type CrawlerConfig struct {
	StartURL        string
	AllowedDomains  []string
	MaxDepth        int
	EnableJS        bool
	EnableScreenshots bool
	CacheEnabled    bool
	BM25Enabled     bool // Placeholder, BM25 is skipped for now
	BM25Query       string // Placeholder
	HeuristicsEnabled bool
	EnableReadability bool // New: Enable Readability
}

// CrawledData stores the extracted information for a URL
type CrawledData struct {
	URL         string
	Markdown      string
	StructuredData map[string]interface{}
	Metadata      map[string]string
	ScreenshotPath string
	RawHTML       string // Optional: For raw data crawling
}

// Crawler struct
type Crawler struct {
	Config      CrawlerConfig
	Cache       map[string]*CrawledData // Simple in-memory cache
	CacheMutex  sync.Mutex
	VisitedURLs map[string]bool
	VisitedMutex sync.Mutex
}

// NewCrawler creates a new Crawler instance
func NewCrawler(config CrawlerConfig) *Crawler {
	return &Crawler{
		Config:      config,
		Cache:       make(map[string]*CrawledData),
		VisitedURLs: make(map[string]bool),
	}
}

// Crawl starts the crawling process
func (c *Crawler) Crawl() (map[string]*CrawledData, error) {
	allCrawledData := make(map[string]*CrawledData)

	collector := colly.NewCollector(
		colly.AllowedDomains(c.Config.AllowedDomains...),
		colly.MaxDepth(c.Config.MaxDepth),
		colly.Async(),
		colly.CacheDir("./.crawler_cache"),
	)

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
		c.VisitedMutex.Lock()
		c.VisitedURLs[r.URL.String()] = true
		c.VisitedMutex.Unlock()
	})

	collector.OnError(func(_ *colly.Response, err error) {
		log.Println("Error:", err)
	})

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		currentURL := e.Request.URL.String()

		if c.Config.CacheEnabled {
			if cachedData := c.getCachedData(currentURL); cachedData != nil {
				fmt.Println("Serving from cache:", currentURL)
				allCrawledData[currentURL] = cachedData
				return
			}
		}

		crawledData := &CrawledData{
			URL:            currentURL,
			StructuredData: make(map[string]interface{}),
			Metadata:       make(map[string]string),
		}

		var doc *goquery.Document
		var err error // Declare err here

		if c.Config.EnableJS {
			dynamicContent, err := c.fetchDynamicContent(currentURL)
			if err != nil {
				log.Printf("Error fetching dynamic content for %s: %v", currentURL, err)
				return
			}
			crawledData.RawHTML = dynamicContent
			doc, err = goquery.NewDocumentFromReader(strings.NewReader(dynamicContent))
			if err != nil {
				log.Printf("Error parsing dynamic HTML for %s: %v", currentURL, err)
				return
			}
		} else {
			crawledData.RawHTML = string(e.Response.Body)
			doc, err = goquery.NewDocumentFromReader(strings.NewReader(crawledData.RawHTML))
			if err != nil {
				log.Printf("Error parsing static HTML for %s: %v", currentURL, err)
				return
			}
		}

		// --- Readability Integration using go-shiori/go-readability ---
		if c.Config.EnableReadability {
			parsedURL, _ := url.Parse(currentURL) // Parse URL for readability
			article, err := readability.FromReader(strings.NewReader(crawledData.RawHTML), parsedURL)
			if err != nil {
				log.Printf("Readability failed for %s: %v. Using raw HTML.", currentURL, err)
				e.DOM = doc.Selection // Fallback to original doc
			} else {
				readabilityDoc, err := goquery.NewDocumentFromReader(strings.NewReader(article.Content))
				if err != nil {
					log.Printf("Error parsing readability output for %s: %v. Using raw HTML.", currentURL, err)
					e.DOM = doc.Selection // Fallback if parsing readability output fails
				} else {
					e.DOM = readabilityDoc.Selection // Use readability's cleaned content
					fmt.Println("Readability applied for:", currentURL)
					crawledData.RawHTML = article.Content // Update RawHTML with cleaned content
				}
			}
		} else {
			e.DOM = doc.Selection // Use the document parsed from raw/dynamic HTML if readability is not enabled
		}


		// 1. Metadata Extraction (Basic)
		crawledData.Metadata["title"] = e.DOM.Find("title").Text()
		crawledData.Metadata["description"], _ = e.DOM.Find("meta[name=description]").Attr("content")

		// 2. Markdown Generation
		markdownContent, references := generateMarkdown(e.DOM, currentURL, c.Config)
		crawledData.Markdown = markdownContent

		if len(references) > 0 {
			crawledData.Markdown += "\n\n**References:**\n"
			for i, ref := range references {
				crawledData.Markdown += fmt.Sprintf("[%d] %s\n", i+1, ref)
			}
		}

		// 3. Structured Data Extraction (Example - Extracting blog post titles and links)
		blogPosts := []map[string]string{}
		e.DOM.Find(".card-body").Each(func(_ int, s *goquery.Selection) {
			title := s.Find("h2.card-title a").Text()
			link, _ := s.Find("h2.card-title a").Attr("href")
			description := s.Find("h4.card-text").Text()
			blogPosts = append(blogPosts, map[string]string{"title": title, "link": e.Request.AbsoluteURL(link), "description": description})
		})
		crawledData.StructuredData["blog_posts"] = blogPosts

		// 4. Screenshot (Optional)
		if c.Config.EnableScreenshots {
			screenshotPath, err := c.captureScreenshot(currentURL)
			if err != nil {
				log.Printf("Error capturing screenshot for %s: %v", currentURL, err)
				return
			} else {
				crawledData.ScreenshotPath = screenshotPath
				fmt.Println("Screenshot saved:", screenshotPath)
			}
		}

		// Cache the data
		if c.Config.CacheEnabled {
			c.cacheData(currentURL, crawledData)
		}
		allCrawledData[currentURL] = crawledData
	})

	collector.Visit(c.Config.StartURL)
	collector.Wait()
	return allCrawledData, nil
}

// getCachedData, cacheData, fetchDynamicContent, captureScreenshot, generateMarkdown, parseSrcset, resolveURL, applyHeuristics - remain the same

// ... (getCachedData, cacheData, fetchDynamicContent, captureScreenshot, generateMarkdown, parseSrcset, resolveURL, applyHeuristics functions are the same as before) ...

// getCachedData retrieves data from cache
func (c *Crawler) getCachedData(urlStr string) *CrawledData {
	c.CacheMutex.Lock()
	defer c.CacheMutex.Unlock()
	return c.Cache[urlStr]
}

// cacheData stores data in cache
func (c *Crawler) cacheData(urlStr string, data *CrawledData) {
	c.CacheMutex.Lock()
	defer c.CacheMutex.Unlock()
	c.Cache[urlStr] = data
}

// fetchDynamicContent uses chromedp to fetch content after JS execution
func (c *Crawler) fetchDynamicContent(urlStr string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var content string
	err := chromedp.Run(ctx,
		chromedp.Navigate(urlStr),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &content, chromedp.ByQuery),
	)
	if err != nil {
		return "", err
	}
	return content, nil
}

// captureScreenshot uses chromedp to capture a screenshot
func (c *Crawler) captureScreenshot(urlStr string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(urlStr),
		chromedp.WaitReady("body"),
		chromedp.CaptureScreenshot(&buf),
	)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("screenshot_%d.png", time.Now().UnixNano())
	filepath := filepath.Join("./screenshots", filename)
	if _, err := os.Stat("./screenshots"); os.IsNotExist(err) {
		os.Mkdir("./screenshots", 0755)
	}

	if err := os.WriteFile(filepath, buf, 0644); err != nil {
		return "", err
	}
	return filepath, nil
}

// generateMarkdown converts HTML to Markdown
func generateMarkdown(selection *goquery.Selection, baseURL string, config CrawlerConfig) (string, []string) {
	var markdownContent strings.Builder
	var references []string

	selection.Find("nav, footer, script, style, noscript").Each(func(_ int, s *goquery.Selection) {
		s.Remove()
	})

	selection.Find(".card-body").Each(func(_ int, cardBody *goquery.Selection) {
		cardBody.Find("h2.card-title a").Each(func(_ int, titleLink *goquery.Selection) {
			title := strings.TrimSpace(titleLink.Text())
			link, _ := titleLink.Attr("href")
			markdownContent.WriteString("## [" + title + "](" + resolveURL(baseURL, link) + ")\n\n")
		})
		cardBody.Find("h4.card-text").Each(func(_ int, desc *goquery.Selection) {
			description := strings.TrimSpace(desc.Text())
			markdownContent.WriteString(description + "\n\n")
		})
	})

	selection.Find("p").Each(func(_ int, p *goquery.Selection) {
		paragraphText := strings.TrimSpace(p.Text())
		if paragraphText != "" {
			markdownContent.WriteString(paragraphText + "\n\n")
		}
	})

	selection.Find("img").Each(func(_ int, img *goquery.Selection) {
		altText, _ := img.Attr("alt")
		src, exists := img.Attr("src")
		if exists {
			absoluteSrc := resolveURL(baseURL, src)
			markdownContent.WriteString(fmt.Sprintf("![%s](%s)\n\n", altText, absoluteSrc))
		}
	})

	selection.Find("picture source").Each(func(_ int, source *goquery.Selection) {
		if srcset, srcsetExists := source.Attr("srcset"); srcsetExists {
			srcsetURLs := parseSrcset(srcset)
			for _, srcsetURL := range srcsetURLs {
				markdownContent.WriteString(fmt.Sprintf("[Image Link](%s)\n\n", resolveURL(baseURL, srcsetURL)))
			}
		}
	})

	selection.Find("audio source, audio").Each(func(_ int, audioElem *goquery.Selection) {
		src, exists := audioElem.Attr("src")
		if exists {
			absoluteSrc := resolveURL(baseURL, src)
			markdownContent.WriteString(fmt.Sprintf("[Audio Link](%s)\n\n", absoluteSrc))
		}
	})

	selection.Find("video source, video").Each(func(_ int, videoElem *goquery.Selection) {
		src, exists := videoElem.Attr("src")
		if exists {
			absoluteSrc := resolveURL(baseURL, src)
			markdownContent.WriteString(fmt.Sprintf("[Video Link](%s)\n\n", absoluteSrc))
		}
	})

	fullMarkdownContent := markdownContent.String()

	if config.HeuristicsEnabled {
		filteredMarkdown := applyHeuristics(fullMarkdownContent)
		markdownContent.Reset()
		markdownContent.WriteString(filteredMarkdown)
		fullMarkdownContent = markdownContent.String()
	}

	markdownContent.Reset()
	markdownContent.WriteString(fullMarkdownContent)

	return markdownContent.String(), references
}

// Helper function to parse srcset attribute
func parseSrcset(srcset string) []string {
	var urls []string
	entries := strings.Split(srcset, ",")
	for _, entry := range entries {
		parts := strings.Fields(strings.TrimSpace(entry))
		if len(parts) > 0 {
			urls = append(urls, strings.TrimSpace(parts[0]))
		}
	}
	return urls
}

// resolveURL resolves relative URLs to absolute URLs
func resolveURL(baseURL string, relativeURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return relativeURL
	}
	rel, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}
	return base.ResolveReference(rel).String()
}

// applyHeuristics applies basic heuristics to filter markdown content
func applyHeuristics(markdownContent string) string {
	var filteredMarkdown strings.Builder
	paragraphs := strings.Split(markdownContent, "\n\n")

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if len(strings.Fields(p)) > 5 {
			filteredMarkdown.WriteString(p + "\n\n")
		}
	}
	return filteredMarkdown.String()
}

func main() {
	app := fiber.New()

	app.Get("/crawl", func(c *fiber.Ctx) error {
		startURL := c.Query("url")
		if startURL == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Please provide a URL as a query parameter, e.g., /crawl?url=https://example.com")
		}

		parsedURL, err := url.ParseRequestURI(startURL)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid URL provided")
		}

		enableReadability := c.QueryBool("readability")

		config := CrawlerConfig{
			StartURL:        startURL,
			AllowedDomains:  []string{parsedURL.Hostname()},
			MaxDepth:        2,
			EnableJS:        false,
			EnableScreenshots: false,
			CacheEnabled:    false,
			HeuristicsEnabled: false,
			EnableReadability: enableReadability,
		}

		crawler := NewCrawler(config)
		crawledDataMap, err := crawler.Crawl()
		if err != nil {
			fiberlog.Errorf("Crawler failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Crawling failed")
		}

		data, ok := crawledDataMap[startURL]
		if !ok {
			return c.Status(fiber.StatusNotFound).SendString("No data crawled for the given URL")
		}

		c.Set("Content-Type", "text/markdown")
		c.Set("Content-Disposition", "inline; filename=\"crawled_content.md\"")

		return c.SendString(data.Markdown)
	})

	fiberlog.Fatal(app.Listen(":3000"))
}