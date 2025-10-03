// SPDX-License-Identifier: MIT
pragma solidity 0.8.26;

interface IPoseidonHasher {
    function hash(uint256 input) external pure returns (uint256 result);

    function identity() external pure returns (uint256);
}

contract PoseidonHasher is IPoseidonHasher {
    uint256 constant Q =
        21888242871839275222246405745257275088548364400416034343698204186575808495617;
    uint256 constant C0 =
        4417881134626180770308697923359573201005643519861877412381846989312604493735;
    uint256 constant C1 =
        5433650512959517612316327474713065966758808864213826738576266661723522780033;
    uint256 constant C2 =
        13641176377184356099764086973022553863760045607496549923679278773208775739952;
    uint256 constant C3 =
        17949713444224994136330421782109149544629237834775211751417461773584374506783;
    uint256 constant C4 =
        13765628375339178273710281891027109699578766420463125835325926111705201856003;
    uint256 constant C5 =
        19179513468172002314585757290678967643352171735526887944518845346318719730387;
    uint256 constant C6 =
        5157412437176756884543472904098424903141745259452875378101256928559722612176;
    uint256 constant C7 =
        535160875740282236955320458485730000677124519901643397458212725410971557409;
    uint256 constant C8 =
        1050793453380762984940163090920066886770841063557081906093018330633089036729;
    uint256 constant C9 =
        10665495010329663932664894101216428400933984666065399374198502106997623173873;
    uint256 constant C10 =
        19965634623406616956648724894636666805991993496469370618546874926025059150737;
    uint256 constant C11 =
        13007250030070838431593222885902415182312449212965120303174723305710127422213;
    uint256 constant C12 =
        16877538715074991604507979123743768693428157847423939051086744213162455276374;
    uint256 constant C13 =
        18211747749504876135588847560312685184956239426147543810126553367063157141465;
    uint256 constant C14 =
        18151553319826126919739798892854572062191241985315767086020821632812331245635;
    uint256 constant C15 =
        19957033149976712666746140949846950406660099037474791840946955175819555930825;
    uint256 constant C16 =
        3469514863538261843186854830917934449567467100548474599735384052339577040841;
    uint256 constant C17 =
        989698510043911779243192466312362856042600749099921773896924315611668507708;
    uint256 constant C18 =
        12568377015646290945235387813564567111330046038050864455358059568128000172201;
    uint256 constant C19 =
        20856104135605479600325529349246932565148587186338606236677138505306779314172;
    uint256 constant C20 =
        8206918720503535523121349917159924938835810381723474192155637697065780938424;
    uint256 constant C21 =
        1309058477013932989380617265069188723120054926187607548493110334522527703566;
    uint256 constant C22 =
        14076116939332667074621703729512195584105250395163383769419390236426287710606;
    uint256 constant C23 =
        10153498892749751942204288991871286290442690932856658983589258153608012428674;
    uint256 constant C24 =
        18202499207234128286137597834010475797175973146805180988367589376893530181575;
    uint256 constant C25 =
        12739388830157083522877690211447248168864006284243907142044329113461613743052;
    uint256 constant C26 =
        15123358710467780770838026754240340042441262572309759635224051333176022613949;
    uint256 constant C27 =
        19925004701844594370904593774447343836015483888496504201331110250494635362184;
    uint256 constant C28 =
        10352416606816998476681131583320899030072315953910679608943150613208329645891;
    uint256 constant C29 =
        10567371822366244361703342347428230537114808440249611395507235283708966113221;
    uint256 constant C30 =
        5635498582763880627392290206431559361272660937399944184533035305989295959602;
    uint256 constant C31 =
        11866432933224219174041051738704352719163271639958083608224676028593315904909;
    uint256 constant C32 =
        5795020705294401441272215064554385591292330721703923167136157291459784140431;
    uint256 constant C33 =
        9482202378699252817564375087302794636287866584767523335624368774856230692758;
    uint256 constant C34 =
        4245237636894546151746468406560945873445548423466753843402086544922216329298;
    uint256 constant C35 =
        12000500941313982757584712677991730019124834399479314697467598397927435905133;
    uint256 constant C36 =
        7596790274058425558167520209857956363736666939016807569082239187494363541787;
    uint256 constant C37 =
        2484867918246116343205467273440098378820186751202461278013576281097918148877;
    uint256 constant C38 =
        18312645949449997391810445935615409295369169383463185688973803378104013950190;
    uint256 constant C39 =
        15320686572748723004980855263301182130424010735782762814513954166519592552733;
    uint256 constant C40 =
        12618438900597948888520621062416758747872180395546164387827245287017031303859;
    uint256 constant C41 =
        17438141672027706116733201008397064011774368832458707512367404736905021019585;
    uint256 constant C42 =
        6374197807230665998865688675365359100400438034755781666913068586172586548950;
    uint256 constant C43 =
        2189398913433273865510950346186699930188746169476472274335177556702504595264;
    uint256 constant C44 =
        6268495580028970231803791523870131137294646402347399003576649137450213034606;
    uint256 constant C45 =
        17896250365994900261202920044129628104272791547990619503076839618914047059275;
    uint256 constant C46 =
        13692156312448722528008862371944543449350293305158722920787736248435893008873;
    uint256 constant C47 =
        15234446864368744483209945022439268713300180233589581910497691316744177619376;
    uint256 constant C48 =
        1572426502623310766593681563281600503979671244997798691029595521622402217227;
    uint256 constant C49 =
        80103447810215150918585162168214870083573048458555897999822831203653996617;
    uint256 constant C50 =
        8228820324013669567851850635126713973797711779951230446503353812192849106342;
    uint256 constant C51 =
        5375851433746509614045812476958526065449377558695752132494533666370449415873;
    uint256 constant C52 =
        12115998939203497346386774317892338270561208357481805380546938146796257365018;
    uint256 constant C53 =
        9764067909645821279940531410531154041386008396840887338272986634350423466622;
    uint256 constant C54 =
        8538708244538850542384936174629541085495830544298260335345008245230827876882;
    uint256 constant C55 =
        7140127896620013355910287215441004676619168261422440177712039790284719613114;
    uint256 constant C56 =
        14297402962228458726038826185823085337698917275385741292940049024977027409762;
    uint256 constant C57 =
        6667115556431351074165934212337261254608231545257434281887966406956835140819;
    uint256 constant C58 =
        20226761165244293291042617464655196752671169026542832236139342122602741090001;
    uint256 constant C59 =
        12038289506489256655759141386763477208196694421666339040483042079632134429119;
    uint256 constant C60 =
        19027757334170818571203982241812412991528769934917288000224335655934473717551;
    uint256 constant C61 =
        16272152964456553579565580463468069884359929612321610357528838696790370074720;
    uint256 constant C62 =
        2500392889689246014710135696485946334448570271481948765283016105301740284071;
    uint256 constant C63 =
        8595254970528530312401637448610398388203855633951264114100575485022581946023;
    uint256 constant C64 =
        11635945688914011450976408058407206367914559009113158286982919675551688078198;
    uint256 constant C65 =
        614739068603482619581328040478536306925147663946742687395148680260956671871;
    uint256 constant C66 =
        18692271780377861570175282183255720350972693125537599213951106550953176268753;
    uint256 constant C67 =
        4987059230784976306647166378298632695585915319042844495357753339378260807164;
    uint256 constant C68 =
        21851403978498723616722415377430107676258664746210815234490134600998983955497;
    uint256 constant C69 =
        9830635451186415300891533983087800047564037813328875992115573428596207326204;
    uint256 constant C70 =
        4842706106434537116860242620706030229206345167233200482994958847436425185478;
    uint256 constant C71 =
        6422235064906823218421386871122109085799298052314922856340127798647926126490;
    uint256 constant C72 =
        4564364104986856861943331689105797031330091877115997069096365671501473357846;
    uint256 constant C73 =
        1944043894089780613038197112872830569538541856657037469098448708685350671343;
    uint256 constant C74 =
        21179865974855950600518216085229498748425990426231530451599322283119880194955;
    uint256 constant C75 =
        14296697761894107574369608843560006996183955751502547883167824879840894933162;
    uint256 constant C76 =
        12274619649702218570450581712439138337725246879938860735460378251639845671898;
    uint256 constant C77 =
        16371396450276899401411886674029075408418848209575273031725505038938314070356;
    uint256 constant C78 =
        3702561221750983937578095019779188631407216522704543451228773892695044653565;
    uint256 constant C79 =
        19721616877735564664624984774636557499099875603996426215495516594530838681980;
    uint256 constant C80 =
        6383350109027696789969911008057747025018308755462287526819231672217685282429;
    uint256 constant C81 =
        20860583956177367265984596617324237471765572961978977333122281041544719622905;
    uint256 constant C82 =
        5766390934595026947545001478457407504285452477687752470140790011329357286275;
    uint256 constant C83 =
        4043175758319898049344746138515323336207420888499903387536875603879441092484;
    uint256 constant C84 =
        15579382179133608217098622223834161692266188678101563820988612253342538956534;
    uint256 constant C85 =
        1864640783252634743892105383926602930909039567065240010338908865509831749824;
    uint256 constant C86 =
        15943719865023133586707144161652035291705809358178262514871056013754142625673;
    uint256 constant C87 =
        2326415993032390211558498780803238091925402878871059708106213703504162832999;
    uint256 constant C88 =
        19995326402773833553207196590622808505547443523750970375738981396588337910289;
    uint256 constant C89 =
        5143583711361588952673350526320181330406047695593201009385718506918735286622;
    uint256 constant C90 =
        15436006486881920976813738625999473183944244531070780793506388892313517319583;
    uint256 constant C91 =
        16660446760173633166698660166238066533278664023818938868110282615200613695857;
    uint256 constant C92 =
        4966065365695755376133119391352131079892396024584848298231004326013366253934;
    uint256 constant C93 =
        20683781957411705574951987677641476019618457561419278856689645563561076926702;
    uint256 constant C94 =
        17280836839165902792086432296371645107551519324565649849400948918605456875699;
    uint256 constant C95 =
        17045635513701208892073056357048619435743564064921155892004135325530808465371;
    uint256 constant C96 =
        17055032967194400710390142791334572297458033582458169295920670679093585707295;
    uint256 constant C97 =
        15727174639569115300068198908071514334002742825679221638729902577962862163505;
    uint256 constant C98 =
        1001755657610446661315902885492677747789366510875120894840818704741370398633;
    uint256 constant C99 =
        18638547332826171619311285502376343504539399518545103511265465604926625041234;
    uint256 constant C100 =
        6751954224763196429755298529194402870632445298969935050224267844020826420799;
    uint256 constant C101 =
        3526747115904224771452549517614107688674036840088422555827581348280834879405;
    uint256 constant C102 =
        15705897908180497062880001271426561999724005008972544196300715293701537574122;
    uint256 constant C103 =
        574386695213920937259007343820417029802510752426579750428758189312416867750;
    uint256 constant C104 =
        15973040855000600860816974646787367136127946402908768408978806375685439868553;
    uint256 constant C105 =
        20934130413948796333037139460875996342810005558806621330680156931816867321122;
    uint256 constant C106 =
        6918585327145564636398173845411579411526758237572034236476079610890705810764;
    uint256 constant C107 =
        14158163500813182062258176233162498241310167509137716527054939926126453647182;
    uint256 constant C108 =
        4164602626597695668474100217150111342272610479949122406544277384862187287433;
    uint256 constant C109 =
        12146526846507496913615390662823936206892812880963914267275606265272996025304;
    uint256 constant C110 =
        10153527926900017763244212043512822363696541810586522108597162891799345289938;
    uint256 constant C111 =
        13564663485965299104296214940873270349072051793008946663855767889066202733588;
    uint256 constant C112 =
        5612449256997576125867742696783020582952387615430650198777254717398552960096;
    uint256 constant C113 =
        12151885480032032868507892738683067544172874895736290365318623681886999930120;
    uint256 constant C114 =
        380452237704664384810613424095477896605414037288009963200982915188629772177;
    uint256 constant C115 =
        9067557551252570188533509616805287919563636482030947363841198066124642069518;
    uint256 constant C116 =
        21280306817619711661335268484199763923870315733198162896599997188206277056900;
    uint256 constant C117 =
        5567165819557297006750252582140767993422097822227408837378089569369734876257;
    uint256 constant C118 =
        10411936321072105429908396649383171465939606386380071222095155850987201580137;
    uint256 constant C119 =
        21338390051413922944780864872652000187403217966653363270851298678606449622266;
    uint256 constant C120 =
        12156296560457833712186127325312904760045212412680904475497938949653569234473;
    uint256 constant C121 =
        4271647814574748734312113971565139132510281260328947438246615707172526380757;
    uint256 constant C122 =
        9061738206062369647211128232833114177054715885442782773131292534862178874950;
    uint256 constant C123 =
        10134551893627587797380445583959894183158393780166496661696555422178052339133;
    uint256 constant C124 =
        8932270237664043612366044102088319242789325050842783721780970129656616386103;
    uint256 constant C125 =
        3339412934966886386194449782756711637636784424032779155216609410591712750636;
    uint256 constant C126 =
        9704903972004596791086522314847373103670545861209569267884026709445485704400;
    uint256 constant C127 =
        17467570179597572575614276429760169990940929887711661192333523245667228809456;
    uint256 constant M00 =
        2910766817845651019878574839501801340070030115151021261302834310722729507541;
    uint256 constant M01 =
        19727366863391167538122140361473584127147630672623100827934084310230022599144;
    uint256 constant M10 =
        5776684794125549462448597414050232243778680302179439492664047328281728356345;
    uint256 constant M11 =
        8348174920934122550483593999453880006756108121341067172388445916328941978568;

    function hash(
        uint256 input
    ) external pure override returns (uint256 result) {
        return _hash(input);
    }

    function _hash(uint256 input) internal pure returns (uint256 result) {
        assembly {
            // Poseidon parameters should be t = 2, RF = 8, RP = 56

            // We load the characteristic
            let q := Q

            // In zerokit implementation, if we pass inp = [a0,a1,..,an] to Poseidon what is effectively hashed is [0,a0,a1,..,an]
            // Note that a sequence of MIX-ARK involves 3 Bn254 field additions before the mulmod happens. Worst case we have a value corresponding to 2*(p-1) which is less than 2^256 and hence doesn't overflow
            //ROUND 0 - FULL
            let s0 := C0
            let s1 := add(input, C1)
            // SBOX
            let t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 1 - FULL
            s0 := add(s0, C2)
            s1 := add(s1, C3)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 2 - FULL
            s0 := add(s0, C4)
            s1 := add(s1, C5)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 3 - FULL
            s0 := add(s0, C6)
            s1 := add(s1, C7)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 4 - PARTIAL
            s0 := add(s0, C8)
            s1 := add(s1, C9)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 5 - PARTIAL
            s0 := add(s0, C10)
            s1 := add(s1, C11)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 6 - PARTIAL
            s0 := add(s0, C12)
            s1 := add(s1, C13)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 7 - PARTIAL
            s0 := add(s0, C14)
            s1 := add(s1, C15)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 8 - PARTIAL
            s0 := add(s0, C16)
            s1 := add(s1, C17)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 9 - PARTIAL
            s0 := add(s0, C18)
            s1 := add(s1, C19)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 10 - PARTIAL
            s0 := add(s0, C20)
            s1 := add(s1, C21)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 11 - PARTIAL
            s0 := add(s0, C22)
            s1 := add(s1, C23)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 12 - PARTIAL
            s0 := add(s0, C24)
            s1 := add(s1, C25)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 13 - PARTIAL
            s0 := add(s0, C26)
            s1 := add(s1, C27)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 14 - PARTIAL
            s0 := add(s0, C28)
            s1 := add(s1, C29)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 15 - PARTIAL
            s0 := add(s0, C30)
            s1 := add(s1, C31)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 16 - PARTIAL
            s0 := add(s0, C32)
            s1 := add(s1, C33)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 17 - PARTIAL
            s0 := add(s0, C34)
            s1 := add(s1, C35)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 18 - PARTIAL
            s0 := add(s0, C36)
            s1 := add(s1, C37)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 19 - PARTIAL
            s0 := add(s0, C38)
            s1 := add(s1, C39)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 20 - PARTIAL
            s0 := add(s0, C40)
            s1 := add(s1, C41)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 21 - PARTIAL
            s0 := add(s0, C42)
            s1 := add(s1, C43)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 22 - PARTIAL
            s0 := add(s0, C44)
            s1 := add(s1, C45)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 23 - PARTIAL
            s0 := add(s0, C46)
            s1 := add(s1, C47)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 24 - PARTIAL
            s0 := add(s0, C48)
            s1 := add(s1, C49)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 25 - PARTIAL
            s0 := add(s0, C50)
            s1 := add(s1, C51)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 26 - PARTIAL
            s0 := add(s0, C52)
            s1 := add(s1, C53)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 27 - PARTIAL
            s0 := add(s0, C54)
            s1 := add(s1, C55)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 28 - PARTIAL
            s0 := add(s0, C56)
            s1 := add(s1, C57)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 29 - PARTIAL
            s0 := add(s0, C58)
            s1 := add(s1, C59)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 30 - PARTIAL
            s0 := add(s0, C60)
            s1 := add(s1, C61)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 31 - PARTIAL
            s0 := add(s0, C62)
            s1 := add(s1, C63)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 32 - PARTIAL
            s0 := add(s0, C64)
            s1 := add(s1, C65)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 33 - PARTIAL
            s0 := add(s0, C66)
            s1 := add(s1, C67)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 34 - PARTIAL
            s0 := add(s0, C68)
            s1 := add(s1, C69)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 35 - PARTIAL
            s0 := add(s0, C70)
            s1 := add(s1, C71)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 36 - PARTIAL
            s0 := add(s0, C72)
            s1 := add(s1, C73)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 37 - PARTIAL
            s0 := add(s0, C74)
            s1 := add(s1, C75)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 38 - PARTIAL
            s0 := add(s0, C76)
            s1 := add(s1, C77)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 39 - PARTIAL
            s0 := add(s0, C78)
            s1 := add(s1, C79)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 40 - PARTIAL
            s0 := add(s0, C80)
            s1 := add(s1, C81)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 41 - PARTIAL
            s0 := add(s0, C82)
            s1 := add(s1, C83)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 42 - PARTIAL
            s0 := add(s0, C84)
            s1 := add(s1, C85)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 43 - PARTIAL
            s0 := add(s0, C86)
            s1 := add(s1, C87)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 44 - PARTIAL
            s0 := add(s0, C88)
            s1 := add(s1, C89)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 45 - PARTIAL
            s0 := add(s0, C90)
            s1 := add(s1, C91)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 46 - PARTIAL
            s0 := add(s0, C92)
            s1 := add(s1, C93)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 47 - PARTIAL
            s0 := add(s0, C94)
            s1 := add(s1, C95)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 48 - PARTIAL
            s0 := add(s0, C96)
            s1 := add(s1, C97)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 49 - PARTIAL
            s0 := add(s0, C98)
            s1 := add(s1, C99)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 50 - PARTIAL
            s0 := add(s0, C100)
            s1 := add(s1, C101)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 51 - PARTIAL
            s0 := add(s0, C102)
            s1 := add(s1, C103)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 52 - PARTIAL
            s0 := add(s0, C104)
            s1 := add(s1, C105)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 53 - PARTIAL
            s0 := add(s0, C106)
            s1 := add(s1, C107)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 54 - PARTIAL
            s0 := add(s0, C108)
            s1 := add(s1, C109)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 55 - PARTIAL
            s0 := add(s0, C110)
            s1 := add(s1, C111)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 56 - PARTIAL
            s0 := add(s0, C112)
            s1 := add(s1, C113)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 57 - PARTIAL
            s0 := add(s0, C114)
            s1 := add(s1, C115)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 58 - PARTIAL
            s0 := add(s0, C116)
            s1 := add(s1, C117)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 59 - PARTIAL
            s0 := add(s0, C118)
            s1 := add(s1, C119)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 60 - FULL
            s0 := add(s0, C120)
            s1 := add(s1, C121)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 61 - FULL
            s0 := add(s0, C122)
            s1 := add(s1, C123)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 62 - FULL
            s0 := add(s0, C124)
            s1 := add(s1, C125)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            t := add(mulmod(s0, M00, q), mulmod(s1, M01, q))
            s1 := add(mulmod(s0, M10, q), mulmod(s1, M11, q))
            s0 := t

            //ROUND 63 - FULL
            s0 := add(s0, C126)
            s1 := add(s1, C127)
            // SBOX
            t := mulmod(s0, s0, q)
            s0 := mulmod(mulmod(t, t, q), s0, q)
            t := mulmod(s1, s1, q)
            s1 := mulmod(mulmod(t, t, q), s1, q)
            // MIX
            s0 := mod(add(mulmod(s0, M00, q), mulmod(s1, M01, q)), q)

            result := s0
        }
    }

    function identity() external pure override returns (uint256) {
        return _identity();
    }

    // The hash of 0
    function _identity() internal pure returns (uint256) {
        return
            0x2a09a9fd93c590c26b91effbb2499f07e8f7aa12e2b4940a3aed2411cb65e11c;
    }
}
