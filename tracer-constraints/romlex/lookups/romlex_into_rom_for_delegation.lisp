(defun  (selector-romlex-into-rom-for-delegation-data)  romlex.COULD_BE_DELEGATION_CODE)

(defun  (bytes-0-3)  (+  (*  (^ 256 2)  (shift  rom.PBCB  0))
                         (*  (^ 256 1)  (shift  rom.PBCB  1))
                         (*  (^ 256 0)  (shift  rom.PBCB  2))
                         ))
(defun  (bytes-3-7)  (+  (*  (^ 256 3)  (shift  rom.PBCB  3))
                         (*  (^ 256 2)  (shift  rom.PBCB  4))
                         (*  (^ 256 1)  (shift  rom.PBCB  5))
                         (*  (^ 256 0)  (shift  rom.PBCB  6))
                         ))
(defun  (bytes-7-16) (+  (*  (^ 256 15)   (shift  rom.PBCB   7 ))
                         (*  (^ 256 14)   (shift  rom.PBCB   8 ))
                         (*  (^ 256 13)   (shift  rom.PBCB   9 ))
                         (*  (^ 256 12)   (shift  rom.PBCB  10 ))
                         (*  (^ 256 11)   (shift  rom.PBCB  11 ))
                         (*  (^ 256 10)   (shift  rom.PBCB  12 ))
                         (*  (^ 256  9)   (shift  rom.PBCB  13 ))
                         (*  (^ 256  8)   (shift  rom.PBCB  14 ))
                         (*  (^ 256  7)   (shift  rom.PBCB  15 ))
                         (*  (^ 256  6)   (shift  rom.PBCB  16 ))
                         (*  (^ 256  5)   (shift  rom.PBCB  17 ))
                         (*  (^ 256  4)   (shift  rom.PBCB  18 ))
                         (*  (^ 256  3)   (shift  rom.PBCB  19 ))
                         (*  (^ 256  2)   (shift  rom.PBCB  20 ))
                         (*  (^ 256  1)   (shift  rom.PBCB  21 ))
                         (*  (^ 256  0)   (shift  rom.PBCB  22 ))
                         ))

(defclookup
  romlex-into-rom-for-delegation-data
  ;; target columns
  (
    (next rom.PC)
    rom.CODE_FRAGMENT_INDEX
    (bytes-0-3)
    (bytes-3-7)
    (bytes-7-16)
  )
  ;; source columns
  (selector-romlex-into-rom-for-delegation-data)
  (
    1
    romlex.CODE_FRAGMENT_INDEX
    romlex.LEADING_THREE_BYTES
    romlex.LEAD_DELEGATION_BYTES
    romlex.TAIL_DELEGATION_BYTES
  )
  )



