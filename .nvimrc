" Disable YAML validation for site.yml
autocmd BufRead,BufNewFile */content/site.yml setlocal filetype=yaml
autocmd FileType yaml if expand('%:t') == 'site.yml' | let b:yaml_schema_ignore = 1 | endif

" Disable specific linters for this file
autocmd BufRead,BufNewFile */content/site.yml let b:ale_linters = []
autocmd BufRead,BufNewFile */content/site.yml let b:coc_diagnostic_disable = 1