-- Mark the hovered row with an arrow (nf-fa-arrow_right) in the padding
-- slot before the icon, instead of a background highlight.
function Entity:padding()
	if self._file.is_hovered and not self._file.in_preview then
		return "\u{F061} "
	end
	return "  "
end

-- Status line: show the mode only outside normal mode, and drop the size
-- and scroll-percentage segments.
function Status:mode()
	local m = self._tab.mode
	if not m.is_select and not m.is_unset then
		return ""
	end
	return ui.Span(" " .. tostring(m):upper() .. " "):style(self:style().main)
end

Status:children_remove(2, Status.LEFT)
Status:children_remove(5, Status.RIGHT)

-- Header: drop the path and keep only the search, filter and find flags.
function Header:cwd()
	local flags = self:flags()
	if flags == "" then
		return ""
	end
	return ui.Span(flags:sub(3, -2)):style(th.mgr.cwd)
end
