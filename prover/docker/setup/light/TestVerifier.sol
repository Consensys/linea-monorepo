

pragma solidity ^0.8.0;
    
import {PlonkVerifier} from './Verifier.sol';


contract TestVerifier {

    using PlonkVerifier for *;

    event PrintBool(bool a);

    struct Proof {
        uint256 proof_l_com_x;
        uint256 proof_l_com_y;
        uint256 proof_r_com_x;
        uint256 proof_r_com_y;
        uint256 proof_o_com_x;
        uint256 proof_o_com_y;

        // h = h_0 + x^{n+2}h_1 + x^{2(n+2)}h_2
        uint256 proof_h_0_x;
        uint256 proof_h_0_y;
        uint256 proof_h_1_x;
        uint256 proof_h_1_y;
        uint256 proof_h_2_x;
        uint256 proof_h_2_y;

        // wire values at zeta
        uint256 proof_l_at_zeta;
        uint256 proof_r_at_zeta;
        uint256 proof_o_at_zeta;

        //uint256[STATE_WIDTH-1] permutation_polynomials_at_zeta; // Sσ1(zeta),Sσ2(zeta)
        uint256 proof_s1_at_zeta; // Sσ1(zeta)
        uint256 proof_s2_at_zeta; // Sσ2(zeta)

        //Bn254.G1Point grand_product_commitment;                 // [z(x)]
        uint256 proof_grand_product_commitment_x;
        uint256 proof_grand_product_commitment_y;

        uint256 proof_grand_product_at_zeta_omega;                    // z(w*zeta)
        uint256 proof_quotient_polynomial_at_zeta;                    // t(zeta)
        uint256 proof_linearised_polynomial_at_zeta;               // r(zeta)

        // Folded proof for the opening of H, linearised poly, l, r, o, s_1, s_2, qcp
        uint256 proof_batch_opening_at_zeta_x;            // [Wzeta]
        uint256 proof_batch_opening_at_zeta_y;

        //Bn254.G1Point opening_at_zeta_omega_proof;      // [Wzeta*omega]
        uint256 proof_opening_at_zeta_omega_x;
        uint256 proof_opening_at_zeta_omega_y;
        
        uint256 proof_openings_selector_commit_api_at_zeta;
        uint256 proof_selector_commit_api_commitment_x;
        uint256 proof_selector_commit_api_commitment_y;
    }

    function get_proof() internal view
    returns (bytes memory)
    {

        Proof memory proof;

        proof.proof_l_com_x = 7624742401655066617302777885579520443775148518530546729814280980688628292274;
        proof.proof_l_com_y = 8008314649620953840365995064029394602254597677673938680986693354913044940445;
        proof.proof_r_com_x = 15884879864496659068567202720458908686853504502377274302723985657666343456176;
        proof.proof_r_com_y = 18091014459594720166950242063957289021958221751651121790779070947297935396026;
        proof.proof_o_com_x = 19131573314687986857464465512053060838809302641733447359696343312495303751474;
        proof.proof_o_com_y = 19749842137321080009834548355330798260192086517877210553235325987443378376216;
        proof.proof_h_0_x = 2093606224943136039236179642016404147024529044856236653589760330818800858961;
        proof.proof_h_0_y = 21806670889775696590296572648936443917052265523033141960833395799656009605711;
        proof.proof_h_1_x = 1947471405439420839159147197980008106915123053292171541132558140049967154152;
        proof.proof_h_1_y = 12265668703393280182513172965341635206702149795073394412420835561876122908740;
        proof.proof_h_2_x = 10418022707940685142655709004259328139499297507617001774421762164912157827814;
        proof.proof_h_2_y = 14016002364587902773120768048785513951010689412813995203648480096044293025131;
        proof.proof_l_at_zeta = 12445719309115933779751501020341734988850425799145883160992986872014071409253;
        proof.proof_r_at_zeta = 17857018084455426205020259963632489248100477923801570584354753860267501824185;
        proof.proof_o_at_zeta = 146142293153491956668167444605799096476870350008613202147436037098306624708;
        proof.proof_s1_at_zeta = 9474035119775642718548378299984874472559923859482427920997604315853403948528;
        proof.proof_s2_at_zeta = 9332837675747061129711589618158270391944119651484863513401954387513220808154;
        proof.proof_grand_product_commitment_x = 19753250111062217149658592779484601711469485218640071619580099722901863408116;
        proof.proof_grand_product_commitment_y = 17697928959763917425502425806090494302823193434264415492710333370558788113240;
        proof.proof_grand_product_at_zeta_omega = 1746724695636529179033854165278294716125541539487092692772989124564103565546;
        proof.proof_quotient_polynomial_at_zeta = 6372884495343243547240624758305064553353776989837805877137721070985678287163;
        proof.proof_linearised_polynomial_at_zeta = 19027712780704464499764543484531807736225144209143191863828396157778172123349;
        proof.proof_batch_opening_at_zeta_x = 162780560311690382870372204977187504412101366667903578014020040428325505996;
        proof.proof_batch_opening_at_zeta_y = 20042822735220107761921677070623826624336664609659421925692791999312319483069;
        proof.proof_opening_at_zeta_omega_x = 7273602848246806004578371949544607713687797194355622701042375221295742865341;
		proof.proof_opening_at_zeta_omega_y = 17110956318700566384289946020513716408461634192611025809636098338161523896461;
        proof.proof_openings_selector_commit_api_at_zeta = 7935256390087037330136231433585708699969692628101864801033489731993292012344   ;
        proof.proof_selector_commit_api_commitment_x = 17759211027173738677060479182253831103514026943473118324159106986163856150772;
        proof.proof_selector_commit_api_commitment_y = 3703273632155718173222088356364857805335661042618413325288863020228347458598;

        bytes memory res;
        res = abi.encodePacked(
            proof.proof_l_com_x,
            proof.proof_l_com_y,
            proof.proof_r_com_x,
            proof.proof_r_com_y,
            proof.proof_o_com_x,
            proof.proof_o_com_y,
            proof.proof_h_0_x,
            proof.proof_h_0_y,
            proof.proof_h_1_x,
            proof.proof_h_1_y,
            proof.proof_h_2_x,
            proof.proof_h_2_y
        );
        res = abi.encodePacked(
            res,
            proof.proof_l_at_zeta,
            proof.proof_r_at_zeta,
            proof.proof_o_at_zeta
        );
        res = abi.encodePacked(
            res,
            proof.proof_s1_at_zeta,
            proof.proof_s2_at_zeta,
            proof.proof_grand_product_commitment_x,
            proof.proof_grand_product_commitment_y,
            proof.proof_grand_product_at_zeta_omega,
            proof.proof_quotient_polynomial_at_zeta,
            proof.proof_linearised_polynomial_at_zeta
        );
        res = abi.encodePacked(
            res,
            proof.proof_batch_opening_at_zeta_x,
            proof.proof_batch_opening_at_zeta_y,
            proof.proof_opening_at_zeta_omega_x,
            proof.proof_opening_at_zeta_omega_y,
            proof.proof_openings_selector_commit_api_at_zeta,
            proof.proof_selector_commit_api_commitment_x,
            proof.proof_selector_commit_api_commitment_y
        );

        return res;
    }

    function test_verifier_go(bytes memory proof, uint256[] memory public_inputs) public {
        bool check_proof = PlonkVerifier.Verify(proof, public_inputs);
        require(check_proof, "verification failed!");
    }

    function test_verifier() public {

        uint256[] memory pi = new uint256[](1);
        
        pi[0] = 0;
        

        bytes memory proof = get_proof();

        bool check_proof = PlonkVerifier.Verify(proof, pi);
        emit PrintBool(check_proof);
        require(check_proof, "verification failed!");
    }

}
