 u s i n g   S y s t e m . C o l l e c t i o n s . G e n e r i c ; 
 u s i n g   S y s t e m . L i n q ; 
 u s i n g   S y s t e m . N e t . M i m e ; 
 u s i n g   M i c r o s o f t . A s p N e t C o r e . H t t p ; 
 u s i n g   M i c r o s o f t . A s p N e t C o r e . M v c ; 
 u s i n g   W e b A p i S a m p l e . M o d e l s ; 
 
 n a m e s p a c e   W e b A p i S a m p l e . C o n t r o l l e r s 
 { 
         # r e g i o n   s n i p p e t _ I n h e r i t 
         [ P r o d u c e s ( M e d i a T y p e N a m e s . A p p l i c a t i o n . J s o n ) ] 
         [ R o u t e ( " [ c o n t r o l l e r ] " ) ] 
         p u b l i c   c l a s s   P e t s C o n t r o l l e r   :   M y C o n t r o l l e r B a s e 
         # e n d r e g i o n 
         { 
                 p r i v a t e   s t a t i c   r e a d o n l y   L i s t < P e t >   _ p e t s I n M e m o r y S t o r e   =   n e w   L i s t < P e t > ( ) ; 
 
                 p u b l i c   P e t s C o n t r o l l e r ( ) 
                 { 
                         i f   ( _ p e t s I n M e m o r y S t o r e . C o u n t   = =   0 ) 
                         { 
                                 _ p e t s I n M e m o r y S t o r e . A d d ( 
                                         n e w   P e t   
                                         {   
                                                 B r e e d   =   " C o l l i e " ,   
                                                 I d   =   1 ,   
                                                 N a m e   =   " F i d o " ,   
                                                 P e t T y p e   =   P e t T y p e . D o g   
                                         } ) ; 
                         } 
                 } 
 
                 [ H t t p G e t ] 
                 p u b l i c   A c t i o n R e s u l t < L i s t < P e t > >   G e t A l l ( )   = >   _ p e t s I n M e m o r y S t o r e ; 
 
                 [ H t t p G e t ( " { i d } " ) ] 
                 [ P r o d u c e s R e s p o n s e T y p e ( S t a t u s C o d e s . S t a t u s 4 0 4 N o t F o u n d ) ] 
                 p u b l i c   A c t i o n R e s u l t < P e t >   G e t B y I d ( i n t   i d ) 
                 { 
                         v a r   p e t   =   _ p e t s I n M e m o r y S t o r e . F i r s t O r D e f a u l t ( p   = >   p . I d   = =   i d ) ; 
 
                         # r e g i o n   s n i p p e t _ P r o b l e m D e t a i l s S t a t u s C o d e 
                         i f   ( p e t   = =   n u l l ) 
                         { 
                                 r e t u r n   N o t F o u n d ( ) ; 
                         } 
                         # e n d r e g i o n 
 
                         r e t u r n   p e t ; 
                 } 
 
                 # r e g i o n   s n i p p e t _ 4 0 0 A n d 2 0 1 
                 [ H t t p P o s t ] 
                 [ P r o d u c e s R e s p o n s e T y p e ( S t a t u s C o d e s . S t a t u s 2 0 1 C r e a t e d ) ] 
                 [ P r o d u c e s R e s p o n s e T y p e ( S t a t u s C o d e s . S t a t u s 4 0 0 B a d R e q u e s t ) ] 
                 p u b l i c   A c t i o n R e s u l t < P e t >   C r e a t e ( P e t   p e t ) 
                 { 
                         p e t . I d   =   _ p e t s I n M e m o r y S t o r e . A n y ( )   ?   
                                           _ p e t s I n M e m o r y S t o r e . M a x ( p   = >   p . I d )   +   1   :   1 ; 
                         _ p e t s I n M e m o r y S t o r e . A d d ( p e t ) ; 
 
                         r e t u r n   C r e a t e d A t A c t i o n ( n a m e o f ( G e t B y I d ) ,   n e w   {   i d   =   p e t . I d   } ,   p e t ) ; 
                 } 
                 # e n d r e g i o n 
         } 
 }