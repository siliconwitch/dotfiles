" jj to escape
inoremap jj <ESC>

" remap bad habits to do nothing
nmap     <Up>    <Nop>
nmap     <Down>  <Nop>
nmap     <Left>  <Nop>
nmap     <Right> <Nop>

" easier navigation in normal / visual / operator pending mode
noremap  K     {
noremap  J     }
noremap  H     ^
noremap  L     $

" make Y consistent with C and D
nnoremap Y y$

" use qq to record, q to stop, Q to play a macro
nnoremap Q @q