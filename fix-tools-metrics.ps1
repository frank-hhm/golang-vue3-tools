# views/tools 下所有 .vue 里 font-size/border-radius/padding/margin(:*) 的 px 字面量替换为 :root 变量
# 严格遵守：字节级 UTF-8 读写，禁止用 Get-Content/Set-Content 默认编码
$ErrorActionPreference = "Stop"
Set-Location "$PSScriptRoot\frontend\src"

$utf8Strict = New-Object System.Text.UTF8Encoding($false, $true)
$utf8Out    = New-Object System.Text.UTF8Encoding($false)

# 字号档位：12 small, 14 default, 18 large, 20 large-x, 32 large-xx
$fontSizeMap = @{
    11 = 'var(--base-size-small)'
    12 = 'var(--base-size-small)'
    13 = 'var(--base-size-small)'
    14 = 'var(--base-size-default)'
    15 = 'var(--base-size-default)'
    16 = 'var(--base-size-default)'
    18 = 'var(--base-size-large)'
    20 = 'var(--base-size-large-x)'
    24 = 'var(--base-size-large-x)'
    26 = 'var(--base-size-large-x)'
    28 = 'var(--base-size-large-xx)'
    30 = 'var(--base-size-large-xx)'
    32 = 'var(--base-size-large-xx)'
    40 = 'var(--base-size-large-xx)'
    48 = 'var(--base-size-large-xx)'
    64 = 'var(--base-size-large-xx)'
}
# 圆角档位：2 small, 6 default, 10 large, 20 large-x
$radiusMap = @{
    1 = 'var(--base-radius-small)'
    2 = 'var(--base-radius-small)'
    3 = 'var(--base-radius-small)'
    4 = 'var(--base-radius-small)'
    5 = 'var(--base-radius-default)'
    6 = 'var(--base-radius-default)'
    7 = 'var(--base-radius-default)'
    8 = 'var(--base-radius-default)'
    9 = 'var(--base-radius-large)'
    10 = 'var(--base-radius-large)'
    12 = 'var(--base-radius-large)'
    14 = 'var(--base-radius-large)'
    16 = 'var(--base-radius-large-x)'
    18 = 'var(--base-radius-large-x)'
    20 = 'var(--base-radius-large-x)'
}
# 间距档位：1-3 → radius-small (2); 4-7 → radius-default (6); 8+ → padding/margin (12)
function Snap-Spacing([int]$px, [bool]$isMargin) {
    $name = if ($isMargin) { 'var(--base-margin)' } else { 'var(--base-padding)' }
    if ($px -le 3) { return 'var(--base-radius-small)' }
    if ($px -le 7) { return 'var(--base-radius-default)' }
    return $name
}

function Replace-Decl($content, [string]$prop, [scriptblock]$snapFn) {
    $rx = [regex]("(?<![\w-])" + [regex]::Escape($prop) + "(\s*:\s*)([^;}\r\n]+)")
    return $rx.Replace($content, {
        param($m)
        $head = $prop + $m.Groups[1].Value
        $val  = $m.Groups[2].Value
        $new = [regex]::Replace($val, '(?<![\d.\w-])(\d+)px\b', {
            param($mm)
            $n = [int]$mm.Groups[1].Value
            $mapped = & $snapFn $n
            if ($mapped) { return $mapped }
            return $mm.Value
        })
        return $head + $new
    })
}

# 需要处理的属性列表（不含 gap，用户明确只说 font-size/radius/padding/margin）
$spacingProps = @('padding','padding-top','padding-right','padding-bottom','padding-left',
                  'margin','margin-top','margin-right','margin-bottom','margin-left')

$totalChanged = 0
Get-ChildItem -Recurse -Include *.vue views\tools | Where-Object { $_.Name -ne 'index-bak.vue' } | ForEach-Object {
    $p = $_.FullName
    $bytes = [System.IO.File]::ReadAllBytes($p)
    # 剥离 UTF-8 BOM（若有）
    $start = 0
    if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) { $start = 3 }
    $content = New-Object byte[] ($bytes.Length - $start)
    [System.Array]::Copy($bytes, $start, $content, 0, $content.Length)
    $text = $utf8Strict.GetString($content)
    $orig = $text

    # font-size
    $text = Replace-Decl $text 'font-size' { param($n) if ($fontSizeMap.ContainsKey($n)) { $fontSizeMap[$n] } else { $null } }
    # border-radius
    $text = Replace-Decl $text 'border-radius' { param($n) if ($radiusMap.ContainsKey($n)) { $radiusMap[$n] } else { $null } }
    # padding / margin（含方向）
    $padSb = [scriptblock]::Create("param(`$n) Snap-Spacing `$n `$false")
    $marSb = [scriptblock]::Create("param(`$n) Snap-Spacing `$n `$true")
    foreach ($prop in $spacingProps) {
        $sb = if ($prop -like 'margin*') { $marSb } else { $padSb }
        $text = Replace-Decl $text $prop $sb
    }

    if ($text -ne $orig) {
        [System.IO.File]::WriteAllBytes($p, $utf8Out.GetBytes($text))
        $totalChanged++
        Write-Host ("updated: " + $p.Substring((Get-Location).Path.Length + 1))
    }
}
Write-Host ("total: $totalChanged files")
