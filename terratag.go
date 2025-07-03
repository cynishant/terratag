package terratag

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cloudyali/terratag/cli"
	"github.com/cloudyali/terratag/internal/cleanup"
	"github.com/cloudyali/terratag/internal/common"
	"github.com/cloudyali/terratag/internal/convert"
	"github.com/cloudyali/terratag/internal/file"
	"github.com/cloudyali/terratag/internal/providers"
	"github.com/cloudyali/terratag/internal/standards"
	"github.com/cloudyali/terratag/internal/tag_keys"
	"github.com/cloudyali/terratag/internal/tagging"
	"github.com/cloudyali/terratag/internal/terraform"
	"github.com/cloudyali/terratag/internal/tfschema"
	"github.com/cloudyali/terratag/internal/utils"
	"github.com/cloudyali/terratag/internal/validation"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

type counters struct {
	totalResources  uint32
	taggedResources uint32
	totalFiles      uint32
	taggedFiles     uint32
}

var pairRegex = regexp.MustCompile(`^([a-zA-Z][\w-]*)=([\w-]+)$`)

var matchWaitGroup sync.WaitGroup

func (c *counters) Add(other counters) {
	atomic.AddUint32(&c.totalResources, other.totalResources)
	atomic.AddUint32(&c.taggedResources, other.taggedResources)
	atomic.AddUint32(&c.totalFiles, other.totalFiles)
	atomic.AddUint32(&c.taggedFiles, other.taggedFiles)
}

// TagLoadingError represents errors that occur during tag loading
type TagLoadingError struct {
	FilePath string
	Cause    string
	Err      error
}

func (e *TagLoadingError) Error() string {
	return fmt.Sprintf("tag loading failed for %s (%s): %v", e.FilePath, e.Cause, e.Err)
}

func (e *TagLoadingError) Unwrap() error {
	return e.Err
}

// loadTagsFromFile loads tags from a tag standardization file and returns them as JSON string
func loadTagsFromFile(filePath string) (string, error) {
	if filePath == "" {
		return "", &TagLoadingError{
			FilePath: filePath,
			Cause:    "empty file path",
			Err:      fmt.Errorf("tag standard file path cannot be empty"),
		}
	}

	// Check if file exists and is readable
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return "", &TagLoadingError{
				FilePath: filePath,
				Cause:    "file not found",
				Err:      fmt.Errorf("tag standard file does not exist: %s", filePath),
			}
		}
		return "", &TagLoadingError{
			FilePath: filePath,
			Cause:    "file access error",
			Err:      fmt.Errorf("cannot access tag standard file: %w", err),
		}
	}

	standard, err := standards.LoadStandard(filePath)
	if err != nil {
		return "", &TagLoadingError{
			FilePath: filePath,
			Cause:    "invalid tag standard format",
			Err:      err,
		}
	}

	// Validate that we have at least some tags defined
	if len(standard.RequiredTags) == 0 && len(standard.OptionalTags) == 0 {
		return "", &TagLoadingError{
			FilePath: filePath,
			Cause:    "no tags defined",
			Err:      fmt.Errorf("tag standard file must define at least one required or optional tag"),
		}
	}

	// Extract tags from the standard file
	// For tagging mode, we'll use required tags and their default values
	tags := make(map[string]string)
	var missingValues []string
	
	// Add required tags with their default values
	for _, tagSpec := range standard.RequiredTags {
		if tagSpec.DefaultValue != "" {
			tags[tagSpec.Key] = tagSpec.DefaultValue
		} else if len(tagSpec.Examples) > 0 {
			tags[tagSpec.Key] = tagSpec.Examples[0]
			log.Printf("[INFO] Using example value '%s' for required tag '%s'", tagSpec.Examples[0], tagSpec.Key)
		} else if len(tagSpec.AllowedValues) > 0 {
			tags[tagSpec.Key] = tagSpec.AllowedValues[0]
			log.Printf("[INFO] Using first allowed value '%s' for required tag '%s'", tagSpec.AllowedValues[0], tagSpec.Key)
		} else {
			// Track tags that need manual configuration
			missingValues = append(missingValues, tagSpec.Key)
			// Use a descriptive placeholder that indicates action needed
			tags[tagSpec.Key] = fmt.Sprintf("CONFIGURE_%s_VALUE", strings.ToUpper(tagSpec.Key))
		}
	}
	
	// Add optional tags with their default values if specified
	for _, tagSpec := range standard.OptionalTags {
		if tagSpec.DefaultValue != "" {
			tags[tagSpec.Key] = tagSpec.DefaultValue
		}
	}

	// Warn about tags that need manual configuration
	if len(missingValues) > 0 {
		log.Printf("[WARN] The following required tags need manual configuration in your tag standard file:")
		for _, key := range missingValues {
			log.Printf("[WARN]   - %s: Add 'default_value', 'examples', or 'allowed_values' to the tag specification", key)
		}
		log.Printf("[WARN] Placeholder values have been used. Update your tag standard file before applying tags.")
	}

	// Convert to JSON string
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return "", &TagLoadingError{
			FilePath: filePath,
			Cause:    "JSON serialization failed",
			Err:      fmt.Errorf("failed to marshal tags to JSON: %w", err),
		}
	}

	log.Printf("[INFO] Successfully loaded %d tags from standard file: %s", len(tags), filePath)
	return string(tagsJSON), nil
}

func Terratag(args cli.Args) error {
	// Create cleanup manager for the operation
	cleanupMgr := cleanup.NewCleanupManager(nil)
	defer func() {
		if err := cleanupMgr.Stop(); err != nil {
			log.Printf("[WARN] Cleanup failed: %v", err)
		}
	}()

	// Handle validation-only mode
	if args.ValidateOnly {
		return validation.ValidateStandards(args)
	}

	// Load tags from the standardization file
	tagsJSON, err := loadTagsFromFile(args.TagsFile)
	if err != nil {
		return fmt.Errorf("failed to load tags from file: %w", err)
	}

	log.Printf("[INFO] Loaded tags from %s: %s", args.TagsFile, tagsJSON)

	if err := terraform.ValidateInitRun(args.Dir, args.Type); err != nil {
		return err
	}

	matches, err := terraform.GetFilePaths(args.Dir, args.Type)
	if err != nil {
		return err
	}

	taggingArgs := &common.TaggingArgs{
		Filter:              args.Filter,
		Skip:                args.Skip,
		Dir:                 args.Dir,
		Tags:                tagsJSON, // Use the loaded tags from file
		Matches:             matches,
		IsSkipTerratagFiles: args.IsSkipTerratagFiles,
		Rename:              args.Rename,
		IACType:             common.IACType(args.Type),
		DefaultToTerraform:  args.DefaultToTerraform,
		KeepExistingTags:    args.KeepExistingTags,
	}

	// Clean up expired provider cache entries (only if cache is enabled)
	if !args.NoProviderCache {
		cacheManager := providers.GetGlobalCacheManager()
		if err := cacheManager.CleanupExpiredEntries(); err != nil {
			log.Printf("[WARN] Failed to cleanup expired cache entries: %v", err)
		}
	}

	// Ensure terraform is initialized if auto-init is enabled
	if args.AutoInit {
		log.Printf("[INFO] Auto-init enabled, ensuring terraform is properly initialized")
		if err := terraform.EnsureInitialized(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); err != nil {
			log.Printf("[WARN] Auto-initialization failed: %v", err)
			log.Printf("[INFO] Continuing with manual initialization check")
		} else {
			log.Printf("[INFO] Terraform initialization verified/completed successfully")
		}
	}

	// Initialize provider schemas before processing files
	if err := tfschema.InitProviderSchemasWithCache(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); err != nil {
		log.Printf("[WARN] Failed to pre-initialize provider schemas: %v", err)
		
		// If auto-init is enabled and schema init failed, try to diagnose and fix
		if args.AutoInit {
			log.Printf("[INFO] Attempting to resolve schema initialization issue with auto-init")
			if initErr := terraform.EnsureInitialized(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); initErr != nil {
				log.Printf("[WARN] Auto-init resolution failed: %v", initErr)
			} else {
				// Retry schema initialization after successful init
				if retryErr := tfschema.InitProviderSchemasWithCache(args.Dir, common.IACType(args.Type), args.DefaultToTerraform, !args.NoProviderCache); retryErr != nil {
					log.Printf("[WARN] Schema initialization still failed after auto-init: %v", retryErr)
				} else {
					log.Printf("[INFO] Schema initialization succeeded after auto-init")
				}
			}
		}
		
		// Continue even if initialization fails, as getResourceSchema will try again on-demand
	}

	// Register cleanup for backup files if they won't be renamed
	if !args.Rename {
		cleanupMgr.AddCleanupHook(func() error {
			return cleanupMgr.CleanupByType(cleanup.ResourceTypeBackupFile)
		})
	}

	counters := tagDirectoryResources(taggingArgs)

	log.Print("[INFO] Summary:")
	log.Print("[INFO] Tagged ", counters.taggedResources, " resource/s (out of ", counters.totalResources, " resource/s processed)")
	log.Print("[INFO] In ", counters.taggedFiles, " file/s (out of ", counters.totalFiles, " file/s processed)")

	return nil
}

func tagDirectoryResources(args *common.TaggingArgs) counters {
	var total counters

	for _, path := range args.Matches {
		if args.IsSkipTerratagFiles && strings.HasSuffix(path, "terratag.tf") {
			log.Print("[INFO] Skipping file ", path, " as it's already tagged")
		} else {
			matchWaitGroup.Add(1)

			go func(path string) {
				defer matchWaitGroup.Done()

				total.Add(counters{
					totalFiles: 1,
				})

				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] failed to process %s due to an exception\n%v", path, r)
					}
				}()

				perFile, err := tagFileResources(path, args)
				if err != nil {
					log.Printf("[ERROR] failed to process %s due to an error\n%v", path, err)

					return
				}

				total.Add(*perFile)
			}(path)
		}
	}

	matchWaitGroup.Wait()

	return total
}

func tagFileResources(path string, args *common.TaggingArgs) (*counters, error) {
	perFileCounters := counters{}

	log.Print("[INFO] Processing file ", path)

	var swappedTagsStrings []string

	hcl, err := file.ReadHCLFile(path)
	if err != nil {
		return nil, err
	}

	filename := file.GetFilename(path)

	hclMap, err := toHclMap(args.Tags)
	if err != nil {
		return nil, err
	}

	terratag := common.TerratagLocal{
		Found: map[string]hclwrite.Tokens{},
		Added: hclMap,
	}

	for _, resource := range hcl.Body().Blocks() {
		switch resource.Type() {
		case "resource":
			log.Print("[INFO] Processing resource ", resource.Labels())

			perFileCounters.totalResources += 1

			matched, err := regexp.MatchString(args.Filter, resource.Labels()[0])
			if err != nil {
				return nil, err
			}

			if !matched {
				log.Print("[INFO] Resource excluded by filter, skipping.", resource.Labels())

				continue
			}

			if args.Skip != "" {
				matched, err = regexp.MatchString(args.Skip, resource.Labels()[0])
				if err != nil {
					return nil, err
				}

				if matched {
					log.Print("[INFO] Resource excluded by skip, skipping.", resource.Labels())

					continue
				}
			}

			isTaggable, err := tfschema.IsTaggable(args.Dir, *resource)
			if err != nil {
				return nil, err
			}

			if isTaggable {
				log.Print("[INFO] Resource taggable, processing...", resource.Labels())

				perFileCounters.taggedResources += 1

				result, err := tagging.TagResource(tagging.TagBlockArgs{
					Filename:         filename,
					Block:            resource,
					Tags:             args.Tags,
					Terratag:         terratag,
					TagId:            providers.GetTagIdByResource(terraform.GetResourceType(*resource)),
					KeepExistingTags: args.KeepExistingTags,
				})
				if err != nil {
					return nil, err
				}

				swappedTagsStrings = append(swappedTagsStrings, result.SwappedTagsStrings...)
			} else {
				log.Print("[INFO] Resource not taggable, skipping.", resource.Labels())
			}
		case "locals":
			// Checks if terratag_added_* exists.
			// If it exists no need to append it again to Terratag file.
			// Instead should override it.
			attributes := resource.Body().Attributes()
			key := tag_keys.GetTerratagAddedKey(filename)

			for attributeKey, attribute := range attributes {
				if attributeKey == key {
					mergedAdded, err := convert.MergeTerratagLocals(attribute, terratag.Added)
					if err != nil {
						return nil, err
					}

					terratag.Added = mergedAdded

					break
				}
			}
		}
	}

	if len(swappedTagsStrings) > 0 {
		convert.AppendLocalsBlock(hcl, filename, terratag)

		text := string(hcl.Bytes())

		swappedTagsStrings = append(swappedTagsStrings, terratag.Added)
		text = convert.UnquoteTagsAttribute(swappedTagsStrings, text)

		if err := file.ReplaceWithTerratagFile(path, text, args.Rename); err != nil {
			return nil, err
		}

		perFileCounters.taggedFiles = 1
	} else {
		log.Print("[INFO] No taggable resources found in file ", path, " - skipping")
	}

	return &perFileCounters, nil
}

func toHclMap(tags string) (string, error) {
	var tagsMap map[string]string

	if err := json.Unmarshal([]byte(tags), &tagsMap); err != nil {
		// If it's not a JSON it might be "key1=value1,key2=value2".
		tagsMap = make(map[string]string)
		pairs := strings.Split(tags, ",")

		for _, pair := range pairs {
			match := pairRegex.FindStringSubmatch(pair)
			if match == nil {
				return "", fmt.Errorf("invalid input tags! must be a valid JSON or pairs of key=value.\nInput: %s", tags)
			}

			tagsMap[match[1]] = match[2]
		}
	}

	keys := utils.SortObjectKeys(tagsMap)

	mapContent := []string{}

	for _, key := range keys {
		mapContent = append(mapContent, "\""+key+"\"="+"\""+tagsMap[key]+"\"")
	}

	return "{" + strings.Join(mapContent, ",") + "}", nil
}
