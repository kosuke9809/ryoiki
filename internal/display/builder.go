package display

import (
	"github.com/kosuke9809/ryoiki/internal/config"
	"github.com/kosuke9809/ryoiki/internal/jj"
)

// BuildWorkspaceInfos merges jj workspace data with ryoiki metadata into WorkspaceInfo slices.
// repoRoot is the repository root path, currentRoot is the current workspace root (used for IsCurrent marking).
func BuildWorkspaceInfos(workspaces []jj.Workspace, metadataMap map[string]config.WorkspaceMetadata, repoRoot, currentRoot string) []WorkspaceInfo {
	var infos []WorkspaceInfo
	for _, w := range workspaces {
		info := WorkspaceInfo{
			Name:        w.Name,
			ChangeID:    w.Target.ChangeID,
			CommitID:    w.Target.CommitID,
			Description: w.Target.Description,
			AuthorName:  w.Target.Author.Name,
			AuthorEmail: w.Target.Author.Email,
		}

		if meta, ok := metadataMap[w.Name]; ok {
			info.Path = meta.Path
			info.Purpose = meta.Purpose
			info.CreatedAt = meta.CreatedAt
			info.UpdatedAt = meta.UpdatedAt
		}

		// Default workspace path is the repo root
		if w.Name == "default" && info.Path == "" {
			info.Path = repoRoot
		}

		// Mark current workspace
		if info.Path == currentRoot {
			info.IsCurrent = true
		}

		infos = append(infos, info)
	}
	return infos
}
